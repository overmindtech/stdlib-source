package network

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/overmindtech/discovery"
)

const TestHTTPTimeout = 3 * time.Second

// TODO: better tests in a controlled environment
func TestHTTPGet(t *testing.T) {
	src := HTTPSource{}

	t.Run("With a valid endpoint", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "https://www.google.com")

		if err != nil {
			t.Fatal(err)
		}

		if i, err := item.Attributes.Get("tls"); err == nil {
			if tlsMap, ok := i.(map[string]interface{}); ok {
				certName := fmt.Sprint(tlsMap["certificate"])

				if matched, _ := regexp.MatchString(`www.google.com \(SHA-1: `, certName); !matched {
					t.Errorf("expected cert name to match www.google.com (SHA-1: , got: %v", certName)
				}
			} else {
				t.Error("expected tls to be map[string]interface{}")
			}
		} else {
			t.Error("expected item to have tls info")
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("With a specified port", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "https://www.google.com:443")

		if err != nil {
			t.Fatal(err)
		}

		var socketFound bool
		var dnsFound bool

		for _, link := range item.LinkedItemRequests {
			switch link.Type {
			case "networksocket":
				socketFound = true

				if link.Query != "www.google.com:443" {
					t.Errorf("expected network socket query to be www.google.com:443, got %v", link.Query)
				}
			case "dns":
				dnsFound = true

				if link.Query != "www.google.com" {
					t.Errorf("expected dns query to be www.google.com, got %v", link.Query)
				}
			}
		}

		if !socketFound {
			t.Error("link to networksocket not found")
		}

		if !dnsFound {
			t.Error("link to dns not found")
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("With an IP", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "https://1.1.1.1:443")

		if err != nil {
			t.Fatal(err)
		}

		var socketFound bool
		var ipFound bool

		for _, link := range item.LinkedItemRequests {
			switch link.Type {
			case "networksocket":
				socketFound = true

				if link.Query != "1.1.1.1:443" {
					t.Errorf("expected network socket query to be 1.1.1.1:443, got %v", link.Query)
				}
			case "ip":
				ipFound = true

				if link.Query != "1.1.1.1" {
					t.Errorf("expected dns query to be 1.1.1.1, got %v", link.Query)
				}
			}
		}

		if !socketFound {
			t.Error("link to networksocket not found")
		}

		if !ipFound {
			t.Error("link to ip not found")
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("With a 404", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "https://httpstat.us/404")

		if err != nil {
			t.Fatal(err)
		}

		var status interface{}

		status, err = item.Attributes.Get("status")

		if err != nil {
			t.Fatal(err)
		}

		if status != float64(404) {
			t.Errorf("expected status to be 404, got: %v", status)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("With a timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		item, err := src.Get(ctx, "global", "http://www.google.com:81/")

		if err == nil {
			t.Errorf("Expected timeout but got %v", item.String())
		}
	})

	t.Run("With a 500 error", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "https://httpstat.us/500")

		if err != nil {
			t.Fatal(err)
		}

		var status interface{}

		status, err = item.Attributes.Get("status")

		if err != nil {
			t.Fatal(err)
		}

		if status != float64(500) {
			t.Errorf("expected status to be 500, got: %v", status)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("With a 301 redirect", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "http://httpstat.us/301")

		if err != nil {
			t.Fatal(err)
		}

		var status interface{}

		status, err = item.Attributes.Get("status")

		if err != nil {
			t.Fatal(err)
		}

		if status != float64(301) {
			t.Errorf("expected status to be 301, got: %v", status)
		}

		if len(item.LinkedItemRequests) == 0 {
			t.Error("expected a linked item to redirected location, got none")
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("With no TLS", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "http://httpstat.us/200")

		if err != nil {
			t.Fatal(err)
		}

		_, err = item.Attributes.Get("tls")

		if err == nil {
			t.Error("Expected to not find TLS info")
		}

		discovery.TestValidateItem(t, item)
	})
}
