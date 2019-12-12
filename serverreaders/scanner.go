package serverreaders

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/genericenricher/enrichers"
)

const (
	defaultPortScanTimeout = time.Second * 3
)

// IPWithPort IP address with a port
type IPWithPort struct {
	IP   net.IP
	Port int
}

// Scanner struct
type Scanner struct {
	Nets          []net.IPNet
	Ports         []int // Ports to scan
	Timeout       time.Duration
	CheckPortOpen bool // Only return servers that have the port open

	serverType  enrichers.ServerType
	readCtx     context.Context
	readCancel  context.CancelFunc
	ipsWithPort chan IPWithPort
}

// NewScanner Create new scanner to scan ips for ports
func NewScanner() *Scanner {
	return &Scanner{Timeout: defaultPortScanTimeout}
}

// SetServerType Set type of server if it is known
func (s *Scanner) SetServerType(t enrichers.ServerType) {
	s.serverType = t
}

// AddIPNet Add IP Network to scan
func (s *Scanner) AddIPNet(n net.IPNet) {
	s.Nets = append(s.Nets, n)
}

// AddPort Add port to scan
func (s *Scanner) AddPort(port int) {
	s.Ports = append(s.Ports, port)
}

// ReadServer Read next server with open port
func (s *Scanner) ReadServer() (genericenricher.Server, error) {
	if s.ipsWithPort == nil {
		s.readCtx, s.readCancel = context.WithCancel(context.Background())
		s.ipsWithPort = s.GetIPsWithPort(s.readCtx)
	}

	for {
		if ipWithPort, ok := <-s.ipsWithPort; ok {
			// Check if server has open port
			if s.CheckPortOpen && !portOpen(ipWithPort.IP, ipWithPort.Port, s.Timeout) {
				// Not of interest, skip
				continue
			}

			// Create genericenricher.Server
			var server genericenricher.Server
			var err error
			if s.serverType == enrichers.Unknown {
				server, err = genericenricher.GetServer(fmt.Sprintf("%s:%d", ipWithPort.IP.String(), ipWithPort.Port))
			} else {
				server, err = genericenricher.GetServerWithType(enrichers.GetConnectionString(ipWithPort.IP, ipWithPort.Port, s.serverType), s.serverType)
			}
			if err != nil {
				// Failed to create server, continue
				continue
			}

			// Return this server
			return server, nil
		}

		// No more ipsWithPort, EOF
		s.readCancel()
		return nil, io.EOF
	}
}

// Close reading of ips
func (s *Scanner) Close() error {
	s.readCancel()
	return nil
}

// Reset back to start of ips
func (s *Scanner) Reset() error {
	s.Close()
	s.ipsWithPort = nil
	return nil
}

// GetIPsWithPort Get all ips with port based on networks to scan and ports to scan
func (s *Scanner) GetIPsWithPort(ctx context.Context) chan IPWithPort {
	ret := make(chan IPWithPort)

	go func() {
		defer close(ret)

		ips := s.GetIPs(ctx)
		for ip := range ips {
			for _, port := range s.Ports {
				select {
				case ret <- IPWithPort{IP: ip, Port: port}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ret
}

// GetIPs Get channel of all IPs in all networks
func (s *Scanner) GetIPs(ctx context.Context) chan net.IP {
	ips := make(chan net.IP)

	go func() {
		defer close(ips)

		for _, n := range s.Nets {
			// Loop through each ip in this CIDR
			baseIP := n.IP
			for ip := baseIP.Mask(n.Mask); n.Contains(ip); inc(ip) {
				new := make(net.IP, len(ip))
				copy(new, ip)
				select {
				case ips <- new:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func portOpen(ip net.IP, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip.String(), port), timeout)

	if err != nil {
		// Check if we have too many connections
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return portOpen(ip, port, timeout)
		}
		return false
	}

	conn.Close()
	return true
}
