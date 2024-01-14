package internet

import (
	"testing"

	"github.com/openrdap/rdap"
	"github.com/openrdap/rdap/bootstrap"
)

func testRdapClient(t *testing.T) *rdap.Client {
	return &rdap.Client{
		Bootstrap: &bootstrap.Client{
			Verbose: func(text string) {
				t.Log(text)
			},
		},
		Verbose: func(text string) {
			t.Log(text)
		},
	}
}
