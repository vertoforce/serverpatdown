# Serverpatdown

[![Go Report Card](https://goreportcard.com/badge/github.com/vertoforce/serverpatdown)](https://goreportcard.com/report/github.com/vertoforce/serverpatdown)
[![Documentation](https://godoc.org/github.com/vertoforce/serverpatdown?status.svg)](https://godoc.org/github.com/vertoforce/serverpatdown)

This package takes a set of servers (or server sources) and searches the _data_ on the servers against a set of regex rules.
Note that unless you call `SetServerDataLimit` this will read **all data** on each server.

This package basically combines `genericenricher` with `multiregex`

## Usage

Example

```go
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
searcher.SetServerDataLimit(1024 * 1024) // 1MB

// Get matches
matchedServers, err := searcher.Process(context.Background())
if err != nil {
    return
}
for matchedServer := range matchedServers {
    fmt.Println(matchedServer.GetConnectString())
}

// Output: http://google.com
```

## TODO

- Add marshal-able state to save and restore sessions
- Make serverreaders threadsafe
- Make multithreaded on processing serverreaders and servers

## Dependencies

- `genericenricher`
- `multiregex`
