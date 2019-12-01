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

const (
	shodanELKQuery = "\"Elastic Indices\""
)

// ShodanReader Finds servers on shodan based on a query.  Implements ServerReader
type ShodanReader struct {
	query            string
	shodanHosts      []*shodan.HostData
	shodanHostsIndex int
	serverType       enrichers.ServerType
}

// NewShodan Create new shodan reader based on a shodan query
func NewShodan(query string, token string, timeout time.Duration) (*ShodanReader, error) {
	s := &ShodanReader{}

	// Make the query
	httpClient := &http.Client{Timeout: timeout}
	client := shodan.NewClient(httpClient, token)

	matchedHosts, err := client.GetHostsForQuery(context.Background(), &shodan.HostQueryOptions{Query: query})
	if err != nil {
		return nil, err
	}
	s.shodanHosts = matchedHosts.Matches

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
		// TODO: Handle this differently?
		return nil, err
	}

	return server, nil
}

func shodanGetConnectionURL(shodanHost *shodan.HostData) string {
	switch shodanHost.Product {
	case "Elastic":
		return fmt.Sprintf("http://%s:%d", shodanHost.IP, shodanHost.Port)
	default:
		return fmt.Sprintf("http://%s:%d", shodanHost.IP, shodanHost.Port)
	}
}
