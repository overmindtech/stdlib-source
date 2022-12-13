package network

import (
	"context"
	"fmt"
	"net"

	"github.com/overmindtech/sdp-go"
)

type SocketSource struct{}

// Type The type of items that this source is capable of finding
func (s *SocketSource) Type() string {
	return "networksocket"
}

// Descriptive name for the source, used in logging and metadata
func (s *SocketSource) Name() string {
	return "stdlib-socket"
}

// List of scopes that this source is capable of find items for. If the
// source supports all scopes the special value `AllScopes` ("*")
// should be used
func (s *SocketSource) Scopes() []string {
	return []string{
		sdp.WILDCARD,
	}
}

// Get Returns a single socket. Note that the query must be in the format
// ip:port. Also in order for the source to return a networksocket in the
// "global" scope, the IP must not be a link-local address i.e. a provate or
// public IP range and not something like 127.0.0.1
func (s *SocketSource) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	var host string
	var port string
	var err error
	var attributes *sdp.ItemAttributes

	attributes, err = sdp.ToAttributes(map[string]interface{}{
		"socket": query,
	})

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	item := sdp.Item{
		Type:            "networksocket",
		UniqueAttribute: "socket",
		Attributes:      attributes,
		Scope:           scope,
	}

	host, port, err = net.SplitHostPort(query)

	if err == nil {
		// Make sure we have been passed a valid IP
		if ip := net.ParseIP(host); ip != nil {
			// Make sure that the IP is valid within this scope
			if scope == "global" && !IsGlobalScopeIP(ip) {
				return nil, &sdp.ItemRequestError{
					ErrorType:   sdp.ItemRequestError_NOTFOUND,
					ErrorString: fmt.Sprintf("%v is not a globally scoped IP. It must be requested with a scope other than global", ip.String()),
					Scope:       scope,
				}
			}

			item.Attributes.Set("ip", host)
			item.Attributes.Set("port", port)

			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ip",
				Method: sdp.RequestMethod_GET,
				Query:  host,
				Scope:  scope,
			})
		} else {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: fmt.Sprintf("%v could not be parsed as an IP address", host),
				Scope:       scope,
			}
		}
	} else {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	return &item, nil
}

// List Is not implemented for HTTP as this would require scanning many
// endpoints or something, doesn't really make sense
func (s *SocketSource) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	return items, nil
}

// Search looks for sockets. It accepts queries in the following formats:
//
// * ip:port
// * dnsName:port
//
// If a DNS name is supplied, that name will be resolved to an IP (or multiple)
// before being returned
func (s *SocketSource) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	var host string
	var port string
	var err error
	var ips []net.IP
	var items []*sdp.Item

	linkedItemRequests := make([]*sdp.ItemRequest, 0)

	host, port, err = net.SplitHostPort(query)

	if err != nil {
		return []*sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	if ip := net.ParseIP(host); ip == nil {
		// If not an IP, try to convert to an IP (or a list of IPs)
		ips, err = net.DefaultResolver.LookupIP(ctx, "ip", host)

		if err != nil {
			return []*sdp.Item{}, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		linkedItemRequests = append(linkedItemRequests, &sdp.ItemRequest{
			Type:   "dns",
			Method: sdp.RequestMethod_GET,
			Query:  host,
			Scope:  "global",
		})
	} else {
		// If it's already an IP just add it our slice
		ips = append(ips, ip)
	}

	// Convert each to a networksocket
	for _, ip := range ips {
		item, err := s.Get(ctx, scope, net.JoinHostPort(ip.String(), port))

		if err != nil {
			return items, err
		}

		// Append DNS query if required
		item.LinkedItemRequests = append(item.LinkedItemRequests, linkedItemRequests...)

		items = append(items, item)
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *SocketSource) Weight() int {
	return 100
}
