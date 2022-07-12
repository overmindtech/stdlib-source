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
	var isGlobalIP bool

	ip = net.ParseIP(query)

	if ip == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("%v is not a valid IP", query),
			Context:     itemContext,
		}
	}

	isGlobalIP = IsGlobalContextIP(ip)

	// If the query was executed with a wildcard, and the context is global, we
	// might was well set it. If it's not then we have no way to determine the
	// context so we need to return an error
	if itemContext == sdp.WILDCARD {
		if isGlobalIP {
			itemContext = "global"
		} else {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: fmt.Sprintf("%v is not a globally-unique IP and therefore could exist in every context. Query with a wildcard does not work for non-global IPs", query),
				Context:     itemContext,
			}
		}
	}

	if itemContext == "global" {
		if !IsGlobalContextIP(ip) {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: fmt.Sprintf("%v is not a valid ip withing the global context. It must be request with some other context", query),
				Context:     itemContext,
			}
		}
	} else {
		// If the context is non-global, ensure that the IP is not globally unique unique
		if IsGlobalContextIP(ip) {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: fmt.Sprintf("%v is a globally-unique IP and therefore only exists in the global context. Note that private IP ranges are also considered 'global' for convenience", query),
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

// IsGlobalContextIP Returns whether or not the IP should be considered valid
// withing the global context according to the following logic:
//
// Non-Global:
//
// * LinkLocalMulticast
// * LinkLocalUnicast
// * InterfaceLocalMulticast
// * Loopback
//
// Global:
//
// * Private
// * Other (All non-reserved addresses)
//
func IsGlobalContextIP(ip net.IP) bool {
	return !(ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsInterfaceLocalMulticast() || ip.IsLoopback())
}
