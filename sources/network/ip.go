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
		// This supports all contexts since there might be local IPs that need
		// to have a different context. E.g. 127.0.0.1 is a different logical
		// address per computer since it referrs to "itself" This means we
		// definitely don't want all thing that reference 127.0.0.1 linked
		// together, only those in the same context
		//
		// TODO: Make a recommendation for what the context should be when
		// looking up an IP in the local range. It's possible that an org could
		// have the address (10.2.56.1) assigned to many devices (hopefully not,
		// but I have seen it happen) and we would therefore want those IPs to
		// have different contexts as they don't refer to the same thing
		sdp.WILDCARD,
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

	if itemContext == "global" {
		// If the context is global, make sure we return an error if it's an IP
		// that isn't globally unique
		var errorString string

		if ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsInterfaceLocalMulticast() {
			errorString = fmt.Sprintf("%v is a link-local address and is therefore not globally unique. It must have a context that is not global", query)
		}
		if ip.IsPrivate() {
			errorString = fmt.Sprintf("%v is a private address and is therefore not globally unique. It must have a context that is not global", query)
		}
		if ip.IsLoopback() {
			errorString = fmt.Sprintf("%v is a loopback address and is therefore not globally unique. It must have a context that is not global", query)
		}

		if errorString != "" {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: errorString,
				Context:     itemContext,
			}
		}
	} else {
		// If the context is non-global, ensure that the IP is not globally unique unique
		if !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsInterfaceLocalMulticast() && !ip.IsLinkLocalMulticast() && !ip.IsLinkLocalUnicast() {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: fmt.Sprintf("%v is a globally-unique IP and therefore only exists in the global context", query),
				Context:     itemContext,
			}
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
