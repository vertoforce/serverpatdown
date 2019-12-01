# Serverpatdown

This package takes a set of servers (or server sources) and searches the _data_ on the servers against a set of regex rules.
Note that unless you call `SetServerDataLimit` this will read **all data** on each server.

This package basically combines `genericenricher` with `multiregex`

## Usage

Example

```go
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
matchedServers, err := searcher.Process(context.Background())
if err != nil {
    return
}
for _, matchedServer := range matchedServers {
    fmt.Println(matchedServer)
}
```

## Dependencies

- `genericenricher`
- `multiregex`
