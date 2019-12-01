package serverreaders

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/vertoforce/genericenricher/enrichers"
)

func TestNewShodan(t *testing.T) {
	shodanReader, err := NewShodan(ShodanELKQuery, os.Getenv("SHODAN_KEY"), time.Second*5)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	shodanReader.SetServerType(enrichers.ELK)

	// Check we can read servers from our query
	serverCount := 0
	for server, err := shodanReader.ReadServer(); err == nil; {
		fmt.Println(server.GetIP())
		serverCount++
		break
	}
	if serverCount == 0 {
		t.Errorf("No servers read")
	}
}
