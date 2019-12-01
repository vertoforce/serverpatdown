// Package serverpatdown takes a set of servers and server sources and searches them against a set of rules
package serverpatdown

import (
	"context"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/multiregex"
)

// ServerReader source of servers
type ServerReader interface {
	ReadServer() (genericenricher.Server, error)
}

// Searcher struct that stores server readers and search rules
type Searcher struct {
	serverReaders   []ServerReader
	servers         []genericenricher.Server
	rules           multiregex.RuleSet
	serverDataLimit int64 // Limit of data to search
}

// AddServerReader Add source of servers
func (searcher *Searcher) AddServerReader(serverReader ServerReader) {
	searcher.serverReaders = append(searcher.serverReaders, serverReader)
}

// AddServer Add a single server
func (searcher *Searcher) AddServer(server genericenricher.Server) {
	searcher.servers = append(searcher.servers, server)
}

// SetServerDataLimit Searcher will read all data on server unless this limit is set
func (searcher *Searcher) SetServerDataLimit(limit int64) {
	searcher.serverDataLimit = limit
}

// AddSearchRule Add search rule
func (searcher *Searcher) AddSearchRule(rule *regexp.Regexp) {
	searcher.rules = append(searcher.rules, rule)
}

// Process Get all servers from readers and search each
func (searcher *Searcher) Process(ctx context.Context) (matchedServers []genericenricher.Server, err error) {
	matchedServers = []genericenricher.Server{}

	// Go through each server
	for _, server := range searcher.servers {
		if searcher.searchServer(ctx, server) {
			matchedServers = append(matchedServers, server)
		}
	}

	// Go through each server reader
	for _, serverReader := range searcher.serverReaders {
		// Read until eof or error
		for server, err := serverReader.ReadServer(); err != io.EOF && err != nil; {
			if searcher.searchServer(ctx, server) {
				matchedServers = append(matchedServers, server)
			}
		}
	}

	return matchedServers, nil
}

func (searcher *Searcher) searchServer(ctx context.Context, server genericenricher.Server) bool {
	// Scan server data (with limit if there is one)
	var matched bool
	if searcher.serverDataLimit == 0 {
		matched = searcher.rules.MatchesRulesReader(ctx, ioutil.NopCloser(server))
	} else {
		matched = searcher.rules.MatchesRulesReader(ctx, ioutil.NopCloser(io.LimitReader(server, searcher.serverDataLimit)))
	}

	server.Close()
	if matched {
		return true
	}
	return false
}
