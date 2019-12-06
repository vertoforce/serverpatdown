// Package serverpatdown takes a set of servers and server sources and searches them against a set of rules
package serverpatdown

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/multiregex"
)

// ServerReader source of servers
type ServerReader interface {
	ReadServer() (genericenricher.Server, error)
	Close() error // Close server reader
}

// Match contains the matching server and regex matches
type Match struct {
	Server  genericenricher.Server
	Matches []string // Matched regexes
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

// AddSearchRulesFromFile Reads file rules (regex rule per line)
func (searcher *Searcher) AddSearchRulesFromFile(filename string) error {
	// Parse regex rules
	inputRegex, err := os.Open(filename)
	if err != nil {
		return errors.New("error opening regex file")
	}
	return searcher.AddSearchRulesFromReader(inputRegex)
}

// AddSearchRulesFromReader Adds regex rules from a reader (regex rule per line)
func (searcher *Searcher) AddSearchRulesFromReader(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		regex, err := regexp.Compile(scanner.Text())
		if err != nil {
			return fmt.Errorf("invalid regex: `%s`", scanner.Text())
		}
		searcher.rules = append(searcher.rules, regex)
	}

	return nil
}

// Process Get all server and search each.  getMatchedData is a parameter to get the
// data the regex rules matched on.  This could miss some matches and will be slower as it won't
// stop on the first match
func (searcher *Searcher) Process(ctx context.Context, getMatchedData bool) (matches chan *Match, err error) {
	matches = make(chan *Match)

	go func() {
		defer close(matches)
		defer func() {
			// Close all readers
			for _, serverReader := range searcher.serverReaders {
				serverReader.Close()
			}
		}()

		// Go through each server
		for _, server := range searcher.servers {
			match := searcher.searchServer(ctx, server, getMatchedData)
			if match.Server != nil {
				// Send match
				select {
				case matches <- match:
				case <-ctx.Done():
					return
				}
			}
		}

		// Go through each server reader
		for _, serverReader := range searcher.serverReaders {
			// Read until eof or error
			for {
				server, err := serverReader.ReadServer()
				if err != nil && err != io.EOF {
					break
				}

				if server != nil {
					match := searcher.searchServer(ctx, server, getMatchedData)
					if match.Server != nil {
						select {
						case matches <- match:
						case <-ctx.Done():
							return
						}
					}
				}

				if err == io.EOF {
					break
				}
			}
			// Close this reader
			serverReader.Close()
		}
	}()

	return matches, nil
}

func (searcher *Searcher) searchServer(ctx context.Context, server genericenricher.Server, getMatchedData bool) *Match {
	match := &Match{}

	if getMatchedData {
		if matches := searcher.getServerMatchedData(ctx, server); len(matches) > 0 {
			match.Server = server
			match.Matches = matches
		}
	} else {
		if searcher.serverMatches(ctx, server) {
			match.Server = server
		}
	}

	return match
}

func (searcher *Searcher) serverMatches(ctx context.Context, server genericenricher.Server) bool {
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

func (searcher *Searcher) getServerMatchedData(ctx context.Context, server genericenricher.Server) []string {
	matchedData := []string{}

	if searcher.serverDataLimit == 0 {
		matches := searcher.rules.GetMatchedDataReader(ctx, server)
		for match := range matches {
			matchedData = append(matchedData, string(match))
		}
	} else {
		matches := searcher.rules.GetMatchedDataReader(ctx, ioutil.NopCloser(io.LimitReader(server, searcher.serverDataLimit)))
		for match := range matches {
			matchedData = append(matchedData, string(match))
		}
	}
	server.Close()

	return matchedData
}
