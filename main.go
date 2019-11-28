// Package eyes takes a set of servers and server sources and searches them against a set of rules
package eyes

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
	GetServer() (genericenricher.Server, error)
}

// Searcher struct that stores server readers and search rules
type Searcher struct {
	serverReaders []ServerReader
	servers       []genericenricher.Server
	rules         multiregex.RuleSet
}

// AddServerReader Add source of servers
func (searcher *Searcher) AddServerReader(serverReader ServerReader) {
	searcher.serverReaders = append(searcher.serverReaders, serverReader)
}

// AddServer Add a single server
func (searcher *Searcher) AddServer(server genericenricher.Server) {
	searcher.servers = append(searcher.servers, server)
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
		for server, err := serverReader.GetServer(); err != io.EOF && err != nil; {
			if searcher.searchServer(ctx, server) {
				matchedServers = append(matchedServers, server)
			}
		}
	}

	return matchedServers, nil
}

func (searcher *Searcher) searchServer(ctx context.Context, server genericenricher.Server) bool {
	// TODO: Change to MatchesRulesReader so it can break early
	matchedRules := searcher.rules.GetMatchedRulesReader(ctx, ioutil.NopCloser(io.LimitReader(server, 100)))
	server.Close()
	if len(matchedRules) > 0 {
		return true
	}
	return false
}
