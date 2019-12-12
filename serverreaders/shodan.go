package serverreaders

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ns3777k/go-shodan/shodan"
	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/genericenricher/enrichers"
)

// Shodan Queries
const (
	ShodanELKQuery = "\"Elastic Indices\""
)

// ShodanReader Finds servers on shodan based on a query.  Implements ServerReader
type ShodanReader struct {
	query            string
	shodanHosts      []*shodan.HostData
	shodanHostsIndex int
	serverType       enrichers.ServerType
	client           *shodan.Client
}

// NewShodan Create new shodan reader based on a shodan query
func NewShodan(ctx context.Context, query string, token string, timeout time.Duration) (*ShodanReader, error) {
	s := &ShodanReader{}
	s.query = query

	// Make the shodan client
	httpClient := &http.Client{Timeout: timeout}
	client := shodan.NewClient(httpClient, token)
	s.client = client

	// Make query
	err := s.Reset()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// SetServerType If the type of servers this will return is already known, set it using this function
func (s *ShodanReader) SetServerType(serverType enrichers.ServerType) {
	s.serverType = serverType
}

// ReadServer Gets next server from Shodan
func (s *ShodanReader) ReadServer() (server genericenricher.Server, err error) {
	// Check if we read all servers
	if s.shodanHostsIndex == len(s.shodanHosts) {
		return nil, io.EOF
	}

	// Return next host we have
	shodanHost := s.shodanHosts[s.shodanHostsIndex]
	s.shodanHostsIndex++

	connectionString := shodanGetConnectionURL(shodanHost)

	// Create server off this
	if s.serverType == enrichers.Unknown {
		// Try to get server type
		server, err = genericenricher.GetServer(connectionString)
	} else {
		// Use known server type
		server, err = genericenricher.GetServerWithType(connectionString, s.serverType)
	}
	if err != nil {
		// TODO: Change this to just get the next server
		return nil, err
	}

	return server, nil
}

// Close shodan server reader
func (s *ShodanReader) Close() error {
	s.shodanHostsIndex = len(s.shodanHosts)
	return nil
}

// Reset make shodan query again and restart processing of hosts
func (s *ShodanReader) Reset() error {
	s.shodanHostsIndex = 0

	// TODO: pagination?
	ctx, cancel := context.WithTimeout(context.Background(), s.client.Client.Timeout)
	matchedHosts, err := s.client.GetHostsForQuery(ctx, &shodan.HostQueryOptions{Query: s.query})
	cancel()
	if err != nil {
		return err
	}
	s.shodanHosts = matchedHosts.Matches

	return nil
}

func shodanGetConnectionURL(shodanHost *shodan.HostData) string {
	switch shodanHost.Product {
	case "Elastic":
		return fmt.Sprintf("http://%s:%d", shodanHost.IP, shodanHost.Port)
	default:
		return fmt.Sprintf("http://%s:%d", shodanHost.IP, shodanHost.Port)
	}
}
