package serverpatdown

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/genericenricher/enrichers"
	"github.com/vertoforce/multiregex"
	"github.com/vertoforce/serverpatdown/serverreaders"
)

// type LocalhostReader struct {
// 	servers []string
// 	serverI int
// }

// func (l *LocalhostReader) GetServer() (genericenricher.Server, error) {
// 	if l.serverI >= len(l.servers) {
// 		return nil, io.EOF
// 	}

// 	return genericenricher.GetServerWithType(l.servers[l.serverI], enrichers.ELK)
// }

func TestProcessWithoutReader(t *testing.T) {
	// Create new searcher
	searcher := &Searcher{}
	searcher.AddSearchRule(multiregex.MatchAll[0])

	// Add a server
	ELKServer, err := genericenricher.GetServerWithType("http://localhost:9200", enrichers.ELK)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	searcher.AddServer(ELKServer)

	// Get matched servers
	matchedServers, err := searcher.Process(context.Background())
	for range matchedServers {
		return
	}
	t.Errorf("Did not match any servers when we should have")
}

func TestProcessWithReader(t *testing.T) {
	// Create new searcher
	searcher := &Searcher{}
	searcher.AddSearchRule(multiregex.MatchAll[0])
	searcher.ReturnNotMatchedServers = true
	searcher.ServerTimeout = time.Millisecond

	// Create server readers
	reader1 := serverreaders.NewScanner()
	reader2 := serverreaders.NewScanner()
	reader1.CheckPortOpen = false
	reader2.CheckPortOpen = false
	reader1.AddPort(80)
	reader2.AddPort(80)
	reader1.SetServerType(enrichers.HTTP)
	reader2.SetServerType(enrichers.HTTP)
	reader1.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPv4Mask(255, 255, 255, 255)})
	reader1.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 2}, Mask: net.IPv4Mask(255, 255, 255, 255)})
	reader1.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 3}, Mask: net.IPv4Mask(255, 255, 255, 255)})
	reader1.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 4}, Mask: net.IPv4Mask(255, 255, 255, 255)})
	reader2.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 5}, Mask: net.IPv4Mask(255, 255, 255, 255)})
	reader2.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 6}, Mask: net.IPv4Mask(255, 255, 255, 255)})
	searcher.AddServerReader(reader1)
	searcher.AddServerReader(reader2)

	// Check depth first
	expectedOrder := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5", "127.0.0.6"}
	searcher.ServerReaderIterationStyle = DepthFirst
	matchedServers, err := searcher.Process(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}
	i := 0
	for matchedServer := range matchedServers {
		if matchedServer.Server.GetIP().String() != expectedOrder[i] {
			t.Errorf("Did not get expected order")
			break
		}
		i++
	}
	if i == 0 {
		t.Errorf("No servers read")
	}

	// Check breadth first
	reader1.Reset()
	reader2.Reset()
	expectedOrder = []string{"127.0.0.1", "127.0.0.5", "127.0.0.2", "127.0.0.6", "127.0.0.3", "127.0.0.4"}
	searcher.ServerReaderIterationStyle = BreadthFirst
	matchedServers, err = searcher.Process(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}
	i = 0
	for matchedServer := range matchedServers {
		if matchedServer.Server.GetIP().String() != expectedOrder[i] {
			t.Errorf("Did not get expected order")
			break
		}
		i++
	}
	if i == 0 {
		t.Errorf("No servers read")
	}
}
