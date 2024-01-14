package internet

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/openrdap/rdap"
	"github.com/overmindtech/sdp-go"
)

func TestEntitySourceSearch(t *testing.T) {
	realUrls := []string{
		"https://rdap.apnic.net/entity/AIC3-AP",
		"https://rdap.apnic.net/entity/IRT-APNICRANDNET-AU",
		"https://rdap.arin.net/registry/entity/GOGL",
	}

	src := &EntitySource{
		Client: testRdapClient(t),
	}

	for _, realUrl := range realUrls {
		t.Run(realUrl, func(t *testing.T) {
			items, err := src.Search(context.Background(), "global", realUrl, false)

			if err != nil {
				t.Fatal(err)
			}

			if len(items) != 1 {
				t.Fatalf("Expected 1 item, got %v", len(items))
			}

			item := items[0]

			err = item.Validate()

			if err != nil {
				t.Error(err)
			}
		})
	}

	t.Run("not found", func(t *testing.T) {
		_, err := src.Search(context.Background(), "global", "https://rdap.apnic.net/entity/NOTFOUND", false)

		if err == nil {
			t.Fatal("Expected error")
		}

		var sdpError *sdp.QueryError

		if ok := errors.As(err, &sdpError); ok {
			if sdpError.ErrorType != sdp.QueryError_NOTFOUND {
				t.Errorf("Expected QueryError_NOTFOUND, got %v", sdpError.ErrorType)
			}
		} else {
			t.Fatalf("Expected QueryError, got %T", err)
		}
	})
}

func TestTest(t *testing.T) {
	request := rdap.Request{
		Type:  rdap.IPRequest,
		Query: "8.8.8.8",
	}

	client := testRdapClient(t)

	response, err := client.Do(&request)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)

	// Get the first entity
	ipNetwork, ok := response.Object.(*rdap.IPNetwork)

	if !ok {
		t.Fatal("Expected IPNetwork")
	}

	if len(ipNetwork.Entities) == 0 {
		t.Fatal("Expected at least one entity")
	}

	exampleEntity := ipNetwork.Entities[0]

	request.Type = rdap.EntityRequest
	request.Query = exampleEntity.Handle

	// Parse the self link
	var selfLink string
	for _, link := range exampleEntity.Links {
		if link.Rel == "self" {
			selfLink = link.Href
		}
	}

	if selfLink == "" {
		t.Fatal("Expected self link")
	}

	// Strip everything after and including /entity since this will be re-added
	// by the library
	selfLink = strings.Split(selfLink, "/entity")[0]

	// Convert to url
	rdapUrl, err := url.Parse(selfLink)

	if err != nil {
		t.Fatal(err)
	}

	request.Server = rdapUrl

	response, err = client.Do(&request)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)
}
