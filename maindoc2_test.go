package serverpatdown

import (
	"context"
	"fmt"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/multiregex"
)

func Example() {
	// Create a searcher object
	searcher := &Searcher{}
	searcher.AddSearchRule(multiregex.MatchAll[0])

	// Add a single server
	server, err := genericenricher.GetServer("http://google.com")
	if err != nil {
		return
	}
	searcher.AddServer(server)

	// Get maches
	matchedServers, err := searcher.Process(context.Background())
	if err != nil {
		return
	}
	for _, matchedServer := range matchedServers {
		fmt.Println(matchedServer)
	}
}
