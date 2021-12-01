package network

import (
	"context"
	"fmt"
	"net"

	"github.com/overmindtech/sdp-go"
)

// IPSource struct on which all methods are registered
type IPSource struct{}

// Type is the type of items that this returns
func (bc *IPSource) Type() string {
	return "ip"
}

// Name Returns the name of the backend
func (bc *IPSource) Name() string {
	return "stdlib-ip"
}

// Weighting of duplicate sources
func (s *IPSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *IPSource) Contexts() []string {
	return []string{
		// DNS entries *should* be globally unique
		"global",
	}
}

// Get gets information about a single IP This expects an IP in a format that
// can be parsed by net.ParseIP() such as "192.0.2.1", "2001:db8::68" or
// "::ffff:192.0.2.1". It returns some useful information about that IP but this
// is all just infmation that is inherent in the IP itself, it doesn't look
// anything up externally
//
// The purpose of this is mainly to provide a node in the graph that many things
// can be linked to, rather than being particularly useful on its own
func (bc *IPSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != "global" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: "IP queries only supported in global context",
			Context:     itemContext,
		}
	}

	var ip net.IP
	var err error
	var attributes *sdp.ItemAttributes

	ip = net.ParseIP(query)

	if ip == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("%v is not a valid IP", query),
			Context:     itemContext,
		}
	}

	attributes, err = sdp.ToAttributes(map[string]interface{}{
		"ip":                      ip.String(),
		"unspecified":             ip.IsUnspecified(),
		"loopback":                ip.IsLoopback(),
		"private":                 ip.IsPrivate(),
		"multicast":               ip.IsMulticast(),
		"interfaceLocalMulticast": ip.IsInterfaceLocalMulticast(),
		"linkLocalMulticast":      ip.IsLinkLocalMulticast(),
		"linkLocalUnicast":        ip.IsLinkLocalUnicast(),
	})

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	return &sdp.Item{
		Type:            "ip",
		UniqueAttribute: "ip",
		Attributes:      attributes,
		Context:         itemContext,
	}, nil
}

// Find Returns an empty list as returning all possible IP addresses would be
// unproductive
func (bc *IPSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != "global" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: "IP queries only supported in global context",
			Context:     itemContext,
		}
	}

	return make([]*sdp.Item, 0), nil
}
