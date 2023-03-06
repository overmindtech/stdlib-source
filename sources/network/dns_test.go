package network

import (
	"context"
	"net"
	"testing"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
)

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
		item, err := src.Get(context.Background(), "global", "one.one.one.one")

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("bad dns entry", func(t *testing.T) {
		_, err := src.Get(context.Background(), "global", "something.does.not.exist.please.testing")

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
		_, err := src.Get(context.Background(), "something.local.test", "something.does.not.exist.please.testing")

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
}
