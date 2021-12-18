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

		if x, err := item.Attributes.Get("ip"); err != nil || x != "173.56.64.34" {
			t.Errorf("expected ip to be 173.56.64.34, got %v%v", err, x)
		}

		if x, err := item.Attributes.Get("port"); err != nil || x != "5432" {
			t.Errorf("expected port to be 5432, got %v%v", err, x)
		}

		if len(item.LinkedItemRequests) != 1 {
			t.Errorf("expected 1 linked item request, got %v", len(item.LinkedItemRequests))
		}
	})

	t.Run("with a DNS name and port", func(t *testing.T) {
		_, err := src.Get(context.Background(), "global", "www.google.com:443")

		if err == nil {
			t.Error("expected error but got <nil>")
		}
	})

	t.Run("with a loopback address", func(t *testing.T) {
		_, err := src.Get(context.Background(), "global", "127.0.0.1:443")

		if err == nil {
			t.Error("expected error but got <nil>")
		}
	})
}

func TestSocketSearch(t *testing.T) {
	src := SocketSource{}

	t.Run("with an IP", func(t *testing.T) {
		items, err := src.Search(context.Background(), "global", "192.168.1.2:443")

		if err != nil {
			t.Error(err)
		}

		discovery.TestValidateItems(t, items)
	})

	t.Run("with a DNS name", func(t *testing.T) {
		items, err := src.Search(context.Background(), "global", "one.one.one.one:53")

		if err != nil {
			t.Error(err)
		}

		discovery.TestValidateItems(t, items)

		if len(items) < 2 {
			t.Fatalf("expected <= 2 items, got %v", len(items))
		}

		firstItem := items[0]

		if len(firstItem.LinkedItemRequests) < 2 {
			t.Fatalf("expected >= 2 linked items requests, got %v", len(firstItem.LinkedItemRequests))
		}
	})

	t.Run("with a bad query", func(t *testing.T) {
		_, err := src.Search(context.Background(), "global", "one.one.one.one")

		if err == nil {
			t.Error("expected error but got <nil>")
		}
	})

	t.Run("with an IPv6 loopback address", func(t *testing.T) {
		_, err := src.Search(context.Background(), "global", "[::1]:443")

		if err == nil {
			t.Error("expected error but got <nil>")
		}
	})

	t.Run("with an IPv6 normal address", func(t *testing.T) {
		items, err := src.Search(context.Background(), "global", "[2a01:4b00:8602:b600::e36d]:443")

		if err != nil {
			t.Error(err)
		}

		discovery.TestValidateItems(t, items)
	})
}
