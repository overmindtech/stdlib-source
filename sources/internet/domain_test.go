package internet

import (
	"context"
	"testing"

	"github.com/overmindtech/sdpcache"
)

func TestDomainSourceGet(t *testing.T) {
	src := &DomainSource{
		Client: testRdapClient(t),
		Cache:  sdpcache.NewCache(),
	}

	items, err := src.Search(context.Background(), "global", "www.google.com", false)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatal("Expected 1 item")
	}

	item := items[0]

	err = item.Validate()

	if err != nil {
		t.Error(err)
	}
}