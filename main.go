// Package eyes takes a set of server sources and searches them against a set of rules
package eyes

import (
	"io"
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
	rules         multiregex.RuleSet
}

// AddServerReader Add source of servers
func (searcher *Searcher) AddServerReader(serverReader ServerReader) {
	searcher.serverReaders = append(searcher.serverReaders, serverReader)
}

// AddSearchRule Add search rule
func (searcher *Searcher) AddSearchRule(rule *regexp.Regexp) {
	searcher.rules = append(searcher.rules, rule)
}

// Process Get all servers from readers and search each
func (searcher *Searcher) Process() error {
	// Go through each server reader
	for _, serverReader := range searcher.serverReaders {
		// Read until eof or error
		for server, err := serverReader.GetServer(); err != io.EOF && err != nil; {
			// TODO: Search server using genericenricher
		}
	}

	return nil
}
