package network

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
)

func TestSearch(t *testing.T) {
	s := DNSSource{
		Servers: []string{
			"1.1.1.1:53",
			"8.8.8.8:53",
		},
	}

	t.Run("with a bad DNS name", func(t *testing.T) {
		_, err := s.Search(context.Background(), "global", "not.real.overmind.tech", false)

		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("with one.one.one.one", func(t *testing.T) {
		items, err := s.Search(context.Background(), "global", "one.one.one.one", false)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 1 {
			t.Errorf("expected 1 item, got %v", len(items))
		}

		// Make sure 1.1.1.1 is in there
		var foundV4 bool
		var foundV6 bool
		for _, item := range items {
			for _, q := range item.LinkedItemQueries {
				if q.Query.Query == "1.1.1.1" {
					foundV4 = true
				}
				if q.Query.Query == "2606:4700:4700::1111" {
					foundV6 = true
				}
			}
		}

		if !foundV4 {
			t.Error("could not find 1.1.1.1 in linked item queries")
		}
		if !foundV6 {
			t.Error("could not find 2606:4700:4700::1111 in linked item queries")
		}

		discovery.TestValidateItems(t, items)
	})

	t.Run("with an IP and therefore reverse DNS", func(t *testing.T) {
		s.ReverseLookup = true
		items, err := s.Search(context.Background(), "global", "1.1.1.1", false)

		if err != nil {
			t.Error(err)
		}

		// Make sure 1.1.1.1 is in there
		var foundV4 bool
		var foundV6 bool
		for _, item := range items {
			for _, q := range item.LinkedItemQueries {
				if q.Query.Query == "1.1.1.1" {
					foundV4 = true
				}
				if q.Query.Query == "2606:4700:4700::1111" {
					foundV6 = true
				}
			}
		}

		if !foundV4 {
			t.Error("could not find 1.1.1.1 in linked item queries")
		}
		if !foundV6 {
			t.Error("could not find 2606:4700:4700::1111 in linked item queries")
		}

		discovery.TestValidateItems(t, items)
	})
}

func TestDnsGet(t *testing.T) {
	var conn net.Conn
	var err error

	// Check that we actually have an inertnet connection, if not there is not
	// point running this test
	conn, err = net.Dial("tcp", "one.one.one.one:443")
	conn.Close()

	if err != nil {
		t.Skip("No internet connection detected")
	}

	src := DNSSource{}

	t.Run("working request", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "one.one.one.one", false)

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("bad dns entry", func(t *testing.T) {
		_, err := src.Get(context.Background(), "global", "something.does.not.exist.please.testing", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if sdpErr, ok := err.(*sdp.QueryError); ok {
			if sdpErr.ErrorType != sdp.QueryError_NOTFOUND {
				t.Errorf("Expected error type to be NOTFOUND, got %v", sdpErr.ErrorType)
			}
		} else {
			t.Errorf("expected error type to be *sdp.QueryError, got %T", err)
		}
	})

	t.Run("bad scope", func(t *testing.T) {
		_, err := src.Get(context.Background(), "something.local.test", "something.does.not.exist.please.testing", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if sdpErr, ok := err.(*sdp.QueryError); ok {
			if sdpErr.ErrorType != sdp.QueryError_NOSCOPE {
				t.Errorf("Expected error type to be NOSCOPE, got %v", sdpErr.ErrorType)
			}
		} else {
			t.Errorf("expected error type to be *sdp.QueryError, got %t", err)
		}
	})

	t.Run("with a CNAME", func(t *testing.T) {
		// When we do a Get on a CNAME, I wan it to work, but only return the
		// first thing
		item, err := src.Get(context.Background(), "global", "www.github.com", false)

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(item)
	})
}
