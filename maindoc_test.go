package serverpatdown

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/vertoforce/genericenricher/enrichers"
	"github.com/vertoforce/multiregex"
	"github.com/vertoforce/serverpatdown/serverreaders"
)

func Example_withServerReader() {
	// Create a searcher object
	searcher := &Searcher{}
	searcher.AddSearchRule(multiregex.MatchAll[0])

	// Create shodan ELK serverreader
	shodanReader, err := serverreaders.NewShodan(serverreaders.ShodanELKQuery, os.Getenv("SHODAN_KEY"), time.Second*5)
	if err != nil {
		return
	}
	shodanReader.SetServerType(enrichers.ELK)

	// Add this reader
	searcher.AddServerReader(shodanReader)

	// Set data limit
	searcher.SetServerDataLimit(1024 * 1024) // 1MB

	// Get matches
	matchedServers, err := searcher.Process(context.Background())
	if err != nil {
		return
	}
	for _, matchedServer := range matchedServers {
		fmt.Println(matchedServer)
	}
}
