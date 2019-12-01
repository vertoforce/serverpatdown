package serverpatdown

import (
	"context"
	"fmt"
	"os"
	"serverpatdown/serverreaders"
	"time"

	"github.com/vertoforce/genericenricher/enrichers"
	"github.com/vertoforce/multiregex"
)

func ExampleSearcher_Process_withServerReader() {
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

	// Get maches
	matchedServers, err := searcher.Process(context.Background())
	if err != nil {
		return
	}
	for _, matchedServer := range matchedServers {
		fmt.Println(matchedServer)
	}
}
