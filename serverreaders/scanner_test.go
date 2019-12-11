package serverreaders

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/vertoforce/genericenricher/enrichers"
)

func TestScanner(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	tsSplit := strings.Split(ts.Listener.Addr().String(), ":")
	tsIP := net.ParseIP(tsSplit[0])
	tsPort, err := strconv.ParseInt(tsSplit[1], 10, 32)
	if err != nil {
		t.Errorf("Error getting testing server IP")
	}

	// Try just getting one server
	s := NewScanner()
	s.AddIPNet(net.IPNet{IP: tsIP, Mask: net.IPMask{255, 255, 255, 255}})
	s.AddPort(int(tsPort))
	s.SetServerType(enrichers.ELK)

	server, err := s.ReadServer()
	if server == nil || server.GetIP().String() != tsIP.String() || server.GetPort() != uint16(tsPort) {
		fmt.Println(server.GetIP())
		fmt.Println(server.GetPort())
		t.Errorf("Did not get correct server")
	}
	if err != nil {
		t.Errorf("Should not have gotten error")
	}

	server, err = s.ReadServer()
	if err != io.EOF {
		t.Errorf("Should have been EOF")
	}

	server, err = s.ReadServer()
	if err != io.EOF {
		t.Errorf("Should have been EOF")
	}
}

func TestGetIPs(t *testing.T) {
	s := NewScanner()

	testNetwork := net.IPNet{IP: net.IP{1, 2, 3, 4}, Mask: net.IPMask{255, 255, 0, 0}}
	s.Nets = []net.IPNet{testNetwork}
	s.Ports = []int{80}

	// Test getting all IPs in the network
	count := 0
	ips := s.GetIPs(context.Background())
	for ip := range ips {
		if ip[0] != 1 || ip[1] != 2 {
			t.Errorf("Incorrect IP: %v", ip)
		}
		count++
	}
	if count != 65536 {
		t.Errorf("Did not loop over all ips")
	}
}
