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

	// Add a single server to scan
	server, err := genericenricher.GetServer("http://google.com")
	if err != nil {
		return
	}
	searcher.AddServer(server)

	// Set data limit
	searcher.SetServerDataLimit(1024 * 1024) // 1MB

	// Get matches
	matchedServers, err := searcher.Process(context.Background(), false)
	if err != nil {
		return
	}
	for matchedServer := range matchedServers {
		fmt.Println(matchedServer)
	}
}
