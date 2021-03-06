package serverpatdown

import (
	"context"
	"fmt"
	"regexp"

	"github.com/vertoforce/genericenricher"
)

func Example() {
	// Create a searcher object
	searcher := NewSearcher()
	searcher.AddSearchRule(regexp.MustCompile(`google`))

	// Add a single server to scan
	server, err := genericenricher.GetServer("http://google.com")
	if err != nil {
		return
	}
	searcher.AddServer(server)

	// Set data limit
	searcher.ServerDataLimit = (1024 * 1024) // 1MB

	// Get matches
	matchedServers, err := searcher.Process(context.Background())
	if err != nil {
		return
	}
	for matchedServer := range matchedServers {
		fmt.Println(matchedServer.Server.GetConnectString())
	}

	// Output: http://google.com
}
