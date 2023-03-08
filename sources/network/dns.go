package network

import (
	"context"
	"fmt"
	"net"

	"github.com/overmindtech/sdp-go"
)

//go:generate docgen ../../doc
// +overmind:type dns
// +overmind:get A DNS entry to look up
// +overmind:list **Not supported**

// +overmind:description Queries DNS records, currently this resolves directly
// to IP addresses rather than CNAMEs etc.

// DNSSource struct on which all methods are registered
type DNSSource struct{}

// Type is the type of items that this returns
func (bc *DNSSource) Type() string {
	return "dns"
}

// Name Returns the name of the backend
func (bc *DNSSource) Name() string {
	return "stdlib-dns"
}

// Weighting of duplicate sources
func (s *DNSSource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *DNSSource) Scopes() []string {
	return []string{
		// DNS entries *should* be globally unique
		"global",
	}
}

// Gets a single item. This expects a DNS name
func (bc *DNSSource) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != "global" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "DNS queries only supported in global scope",
			Scope:       scope,
		}
	}

	// Check for IP addresses and do nothing
	if net.ParseIP(query) != nil {
		return &sdp.Item{}, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: fmt.Sprintf("%v is already an IP address, no DNS entry will be found", query),
		}
	}

	var err error
	var i sdp.Item
	var ips []string
	var ipsInterface []interface{}

	ips, err = net.DefaultResolver.LookupHost(ctx, query)

	for _, ip := range ips {
		// Link this to a "global" IP object
		i.LinkedItemQueries = append(i.LinkedItemQueries, &sdp.Query{
			Scope:  "global",
			Method: sdp.RequestMethod_GET,
			Query:  ip,
			Type:   "ip",
		})

		// Convert IPs to a slice of interfaces since this is what protobuf needs in
		// order to be able to convert correctly
		ipsInterface = append(ipsInterface, ip)
	}

	if err != nil {
		// Check if this was a no such host error, if this is the case we want
		// to return a nice "Not Found" error
		if netErr, ok := err.(*net.DNSError); ok {
			if netErr.IsNotFound {
				return &i, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: err.Error(),
					Scope:       scope,
				}
			}
		}

		return &i, err
	}

	i.Type = "dns"
	i.UniqueAttribute = "name"
	i.Scope = "global"
	i.Attributes, err = sdp.ToAttributes(map[string]interface{}{
		"name": query,
		"ips":  ipsInterface,
	})

	return &i, err
}

// List calls back to the ListFunction to find all items
func (bc *DNSSource) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != "global" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "DNS queries only supported in global scope",
			Scope:       scope,
		}
	}

	return make([]*sdp.Item, 0), nil
}
