package network

import (
	"context"
	"regexp"
	"testing"
)

func TestIPGet(t *testing.T) {
	src := IPSource{}

	t.Run("with ipv4 address", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "192.168.1.27")

		if err != nil {
			t.Fatal(err)
		}

		if private, err := item.Attributes.Get("private"); err == nil {
			if private != true {
				t.Error("Expected itemattributes.private to be true")
			}
		} else {
			t.Error("could not find 'private' attribute")
		}
	})

	t.Run("with ipv6 address", func(t *testing.T) {
		item, err := src.Get(context.Background(), "global", "2a01:4b00:8602:b600:5523:ce8d:dafc:3243")

		if err != nil {
			t.Fatal(err)
		}

		if private, err := item.Attributes.Get("private"); err == nil {
			if private != false {
				t.Error("Expected itemattributes.private to be false")
			}
		} else {
			t.Error("could not find 'private' attribute")
		}
	})

	t.Run("with invalid address", func(t *testing.T) {
		_, err := src.Get(context.Background(), "global", "this is not valid")

		if err == nil {
			t.Error("expected error")
		} else {
			if matched, _ := regexp.MatchString("this is not valid", err.Error()); !matched {
				t.Errorf("expected error to contain 'this is not valid', got: %v", err)
			}
		}
	})
}
