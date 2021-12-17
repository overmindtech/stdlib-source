package network

import (
	"context"
	"testing"

	"github.com/overmindtech/discovery"
)

func TestSocketGet(t *testing.T) {
	src := SocketSource{}

	t.Run("with a DNS name and no port", func(t *testing.T) {
		_, err := src.Get(context.Background(), "global", "foo.bar.com")

		if err == nil {
			t.Error("expected error but got <nil>")
		}
	})

	t.Run("with an IP and port", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "173.56.64.34:5432")

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)

		if x, err := item.Attributes.Get("host"); err != nil || x != "173.56.64.34" {
			t.Errorf("expected host to be 173.56.64.34, got %v%v", err, x)
		}

		if x, err := item.Attributes.Get("port"); err != nil || x != "5432" {
			t.Errorf("expected port to be 5432, got %v%v", err, x)
		}

		if len(item.LinkedItemRequests) != 1 {
			t.Errorf("expected 1 linked item request, got %v", len(item.LinkedItemRequests))
		}
	})

	t.Run("with a DNS name and port", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "www.google.com:443")

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)
	})
}
