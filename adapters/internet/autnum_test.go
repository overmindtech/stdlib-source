package internet

import (
	"context"
	"testing"

	"github.com/openrdap/rdap"
	"github.com/overmindtech/sdpcache"
)

func TestASNAdapterGet(t *testing.T) {
	t.Parallel()

	src := &ASNAdapter{
		ClientFac: func() *rdap.Client { return testRdapClient(t) },
		Cache:     sdpcache.NewCache(),
	}

	item, err := src.Get(context.Background(), "global", "AS15169", false)

	if err != nil {
		t.Fatal(err)
	}

	err = item.Validate()

	if err != nil {
		t.Error(err)
	}
}
