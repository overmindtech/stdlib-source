package internet

import (
	"context"
	"testing"
)

func TestASNSourceGet(t *testing.T) {
	src := &ASNSource{
		Client: testRdapClient(t),
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
