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
	"time"

	"github.com/vertoforce/genericenricher"
	"github.com/vertoforce/multiregex"
)

// IterationStyle How to iterate over the readers, breadth first or depth first
type IterationStyle int

const (
	// BreadthFirst Iterate over each ServerReader reading a server round robin style
	BreadthFirst IterationStyle = iota
	// DepthFirst Read all servers from a ServerReader before moving on to the next reader
	DepthFirst

	defaultServerTimeout = time.Second * 4
)

// ServerReader Source of servers, should return EOF on each read after EOF
// TODO: Convention on what to do with error on creation of server
type ServerReader interface {
	ReadServer() (genericenricher.Server, error)
	Close() error // Close server reader
	Reset() error // Reset to start reading servers again
}

// Match contains the matching server and regex matches
type Match struct {
	Matched bool
	Server  genericenricher.Server
	Matches []multiregex.Match // Matched regexes
}

// Searcher struct that stores server readers and search rules
type Searcher struct {
	// Get the data the regex rules matched on (Match.Matches). This could miss some matches and will be slower as it won't stop on the first match.
	GetMatchedData bool
	// Return servers that did not match (with Match.Matched=false) for logging or progress tracking
	ReturnNotMatchedServers bool
	// Limit of data to read on each server
	ServerDataLimit int64
	// Style of iterating over readers (breadth first or depth first)
	ServerReaderIterationStyle IterationStyle
	// Timeout to connect to each server
	ServerTimeout time.Duration

	serverReaders []ServerReader
	servers       []genericenricher.Server
	rules         multiregex.RuleSet
}

func NewSearcher() *Searcher {
	s := &Searcher{ServerTimeout: defaultServerTimeout}
	return s
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
		searcher.AddSearchRule(regex)
	}

	return nil
}

// Process Get all servers and search each.
// It first scans all single servers added, then goes depth/breadth for each server reader
func (searcher *Searcher) Process(ctx context.Context) (matches chan *Match, err error) {
	matches = make(chan *Match)

	go func() {
		defer close(matches)
		defer func() {
			// Close all readers
			for _, serverReader := range searcher.serverReaders {
				serverReader.Close()
			}
		}()

		// Process each server
		for _, server := range searcher.servers {
			searcher.processServer(ctx, server, matches)
		}

		// Process readers
		if searcher.ServerReaderIterationStyle == BreadthFirst {
			for {
				// Keep looping over each reader until we've finished them all
				finishedReaders := 0
				for _, serverReader := range searcher.serverReaders {
					// Read and process a server
					if !searcher.processAServerReaderServer(ctx, serverReader, matches) {
						// Done reading this
						finishedReaders++
					}
				}
				if finishedReaders == len(searcher.serverReaders) {
					break
				}
			}
		} else if searcher.ServerReaderIterationStyle == DepthFirst {
			for _, serverReader := range searcher.serverReaders {
				// Process all servers in this reader
				for searcher.processAServerReaderServer(ctx, serverReader, matches) {
				}
			}
		} else {
			return
		}

	}()

	return matches, nil
}

// processAServerReaderServer Read single server from ServerReader and process, returns true if there is more to be read
func (searcher *Searcher) processAServerReaderServer(ctx context.Context, serverReader ServerReader, matches chan *Match) bool {
	server, err := serverReader.ReadServer()
	if err != nil && err != io.EOF {
		// Close this reader
		serverReader.Close()
		return false
	}

	if server != nil {
		searcher.processServer(ctx, server, matches)
	}

	if err == io.EOF {
		// Close this reader
		serverReader.Close()
		return false
	}

	return true
}

// processServer Given a server send the associated match
func (searcher *Searcher) processServer(ctx context.Context, server genericenricher.Server, matches chan *Match) {
	match := searcher.searchServer(ctx, server, searcher.GetMatchedData)
	if match.Matched || searcher.ReturnNotMatchedServers {
		// Send match
		select {
		case matches <- match:
		case <-ctx.Done():
			return
		}
	}
}

// searchServer Search a server and return the match
func (searcher *Searcher) searchServer(ctx context.Context, server genericenricher.Server, getMatchedData bool) *Match {
	match := &Match{}
	match.Server = server
	match.Matched = false

	// Check if we can connect
	c, cancel := context.WithTimeout(ctx, searcher.ServerTimeout)
	err := server.Connect(c)
	if err != nil {
		cancel()
		return match
	}

	// Create new reader if we have a limit
	var serverReader io.ReadCloser
	if searcher.ServerDataLimit == 0 {
		serverReader = server
	} else {
		serverReader = ioutil.NopCloser(io.LimitReader(server, searcher.ServerDataLimit))
	}

	if searcher.GetMatchedData {
		// Get the matched data
		matchesChan := searcher.rules.GetMatchedDataReader(ctx, serverReader)

		// Read all matched rules and data
		ms := []multiregex.Match{}
		for m := range matchesChan {
			ms = append(ms, m)
		}
		match.Matches = ms

		// Check if we got any
		if len(match.Matches) > 0 {
			match.Matched = true
		}
	} else {
		// Check if we match
		if searcher.rules.MatchesRulesReader(ctx, serverReader) {
			match.Matched = true
		}
	}

	// Cancel connection to server
	cancel()

	return match
}
