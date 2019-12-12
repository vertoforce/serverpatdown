package serverreaders

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/vertoforce/genericenricher/enrichers"
)

func TestNewShodan(t *testing.T) {
	shodanReader, err := NewShodan(context.Background(), ShodanELKQuery, os.Getenv("SHODAN_KEY"), time.Second*5)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	shodanReader.SetServerType(enrichers.ELK)

	// Check we can read servers from our query
	serverCount := 0
	for {
		server, err := shodanReader.ReadServer()
		if err != nil {
			break
		}
		fmt.Println(server.GetIP())
		serverCount++
	}
	if serverCount == 0 {
		t.Errorf("No servers read")
	}

	// Check EOFs
	_, err = shodanReader.ReadServer()
	if err != io.EOF {
		t.Errorf("Should have been eof")
	}
	_, err = shodanReader.ReadServer()
	if err != io.EOF {
		t.Errorf("Should have been eof")
	}

	// Check EOFs after close
	shodanReader.Close()

	_, err = shodanReader.ReadServer()
	if err != io.EOF {
		t.Errorf("Should have been eof")
	}
	_, err = shodanReader.ReadServer()
	if err != io.EOF {
		t.Errorf("Should have been eof")
	}

	// Check reset
	shodanReader.Reset()
	_, err = shodanReader.ReadServer()
	if err != nil {
		t.Errorf(err.Error())
	}
}
