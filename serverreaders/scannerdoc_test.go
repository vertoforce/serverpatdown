package serverreaders

import (
	"fmt"
	"net"

	"github.com/vertoforce/genericenricher/enrichers"
)

func ExampleNewScanner() {
	s := NewScanner()
	s.AddIPNet(net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}})
	s.AddPort(9200)
	s.SetServerType(enrichers.ELK)

	server, _ := s.ReadServer()
	fmt.Println(server.GetIP().String())

	// Output: 127.0.0.1
}
