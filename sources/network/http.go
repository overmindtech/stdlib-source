package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/overmindtech/sdp-go"
)

const USER_AGENT_VERSION = "0.1"

type HTTPSource struct {
	// If you were writing a source that needed to store some state or config,
	// you would store that config in here. Since this is just static and
	// doesn't connect to any other systems that might warrant configuration,
	// we'll leave this blank
}

// Type The type of items that this source is capable of finding
func (s *HTTPSource) Type() string {
	return "http"
}

// Descriptive name for the source, used in logging and metadata
func (s *HTTPSource) Name() string {
	return "stdlib-http"
}

// List of contexts that this source is capable of find items for. If the
// source supports all contexts the special value `AllContexts` ("*")
// should be used
func (s *HTTPSource) Contexts() []string {
	return []string{
		"global", // This is a reserved word meaning that the items should be considered globally unique
	}
}

// Get Get a single item with a given context and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *HTTPSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != "global" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: "http is are only supported in the 'global' context",
			Context:     itemContext,
		}
	}

	// Create a client that skips TLS verification since we will want to get the
	// details of the TLS connection rather than stop if it's not trusted. Since
	// we are only running a HEAD request this is unlikely to be a problem
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: tr,
		// Don't follow redirects, just return teh status code directly
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", query, http.NoBody)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	req.Header.Add("User-Agent", fmt.Sprintf("Overmind/%v (%v/%v)", USER_AGENT_VERSION, runtime.GOOS, runtime.GOARCH))
	req.Header.Add("Accept", "*/*")

	var res *http.Response

	res, err = client.Do(req)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	// Convert headers from map[string][]string to map[string]string. This means
	// that headers that were returned many times will end up with their values
	// comma-separated
	headersMap := make(map[string]string)
	for header, values := range res.Header {
		headersMap[header] = strings.Join(values, ", ")
	}

	// Convert the attributes from a golang map, to the structure required for
	// the SDP protocol
	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"name":             query,
		"status":           res.StatusCode,
		"statusString":     res.Status,
		"proto":            res.Proto,
		"headers":          headersMap,
		"transferEncoding": res.Request.TransferEncoding,
	})

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	if tlsState := res.TLS; tlsState != nil {
		var version string

		// Extract TLS version as a string
		switch tlsState.Version {
		case tls.VersionTLS10:
			version = "TLSv1.0"
		case tls.VersionTLS11:
			version = "TLSv1.1"
		case tls.VersionTLS12:
			version = "TLSv1.2"
		case tls.VersionTLS13:
			version = "TLSv1.3"
		//lint:ignore SA1019 We are just *checking* SSLv3, not using it
		case tls.VersionSSL30:
			version = "SSLv3"
		default:
			version = "unknown"
		}

		attributes.Set("tls", map[string]interface{}{
			"version":     version,
			"certificate": CertToName(tlsState.PeerCertificates[0]),
			"serverName":  tlsState.ServerName,
		})
	}

	item := sdp.Item{
		Type:               "http",
		UniqueAttribute:    "name",
		Attributes:         attributes,
		Context:            "global",
		LinkedItemRequests: []*sdp.ItemRequest{
			// TODO: Add linked item request for the certificate once we've created that source
			// {
			// 	Type: "certificate",
			// 	Method: sdp.RequestMethod_SEARCH,
			// }
		},
	}

	// Detect redirect and add a linked item for the redirect target
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		if loc := res.Header.Get("Location"); loc != "" {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:    "http",
				Method:  sdp.RequestMethod_GET,
				Query:   loc,
				Context: itemContext,
			})
		}
	}

	return &item, nil
}

// Find Is not implemented for HTTP as this would require scanning many
// endpoints or something, doesn't really make sense
func (s *HTTPSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *HTTPSource) Weight() int {
	return 100
}
