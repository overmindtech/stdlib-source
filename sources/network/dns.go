package network

import (
	"context"
	"fmt"
	"net"

	"github.com/overmindtech/sdp-go"
)

// DNSSource struct on which all methods are registered
type DNSSource struct{}

// Type is the type of items that this returns
func (bc *DNSSource) Type() string {
	return "dns"
}

// Name Returns the name of the backend
func (bc *DNSSource) Name() string {
	return "network"
}

// Weighting of duplicate sources
func (s *DNSSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *DNSSource) Contexts() []string {
	return []string{
		// DNS entries *should* be globally unique
		"global",
	}
}

// Get calls back to the GetFunction to get a single item. This expects a DNS
// name
func (bc *DNSSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != "global" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: "DNS queries only supported in global context",
			Context:     itemContext,
		}
	}

	// Check for IP addresses and do nothing
	if net.ParseIP(query) != nil {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("%v is already an IP address, no DNS entry will be found", query),
		}
	}

	var err error
	var i sdp.Item
	var ips []string
	var ipsInterface []interface{}

	ips, err = net.LookupHost(query)

	for _, ip := range ips {
		// Link this to a "global" IP object
		i.LinkedItemRequests = append(i.LinkedItemRequests, &sdp.ItemRequest{
			Context: "global",
			Method:  sdp.RequestMethod_GET,
			Query:   ip,
			Type:    "ip",
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
				return &i, &sdp.ItemRequestError{
					ErrorType:   sdp.ItemRequestError_NOTFOUND,
					ErrorString: err.Error(),
					Context:     itemContext,
				}
			}
		}

		return &i, err
	}

	i.Type = "dns"
	i.UniqueAttribute = "name"
	i.Context = "global"
	i.Attributes, err = sdp.ToAttributes(map[string]interface{}{
		"name": query,
		"ips":  ipsInterface,
	})

	return &i, err
}

// Find calls back to the FindFunction to find all items
func (bc *DNSSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != "global" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: "DNS queries only supported in global context",
			Context:     itemContext,
		}
	}

	return make([]*sdp.Item, 0), nil
}
