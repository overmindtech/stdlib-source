package internet

import (
	"context"
	"testing"

	"github.com/openrdap/rdap"
)

func TestIpNetworkSourceSearch(t *testing.T) {
	src := &IPNetworkSource{
		Client: &rdap.Client{},
	}

	items, err := src.Search(context.Background(), "global", "1.1.1.1", false)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %v", len(items))
	}

	item := items[0]

	if item.UniqueAttributeValue() != "1.1.1.0 - 1.1.1.255" {
		t.Errorf("Expected unique attribute value to be 1.1.1.0 - 1.1.1.0 - 1.1.1.255, got %v", item.UniqueAttributeValue())
	}

	if len(item.LinkedItemQueries) != 3 {
		t.Errorf("Expected 3 linked items, got %v", len(item.LinkedItemQueries))
	}
}
