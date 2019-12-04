package serverpatdown

import (
	"context"
	"testing"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/genericenricher/enrichers"
	"github.com/vertoforce/multiregex"
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
	matchedServers, err := searcher.Process(context.Background(), false)
	for range matchedServers {
		return
	}
	t.Errorf("Did not match any servers when we should have")
}

// TODO: TestProcessWithReader
