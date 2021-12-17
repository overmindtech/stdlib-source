package network

import (
	"context"
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

// List of contexts that this source is capable of find items for. If the
// source supports all contexts the special value `AllContexts` ("*")
// should be used
func (s *SocketSource) Contexts() []string {
	return []string{
		sdp.WILDCARD,
	}
}

// Get Returns a single socket. Note that the query must be in the format host:port
func (s *SocketSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
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
			Context:     itemContext,
		}
	}

	item := sdp.Item{
		Type:            "networksocket",
		UniqueAttribute: "socket",
		Attributes:      attributes,
		Context:         itemContext,
	}

	host, port, err = net.SplitHostPort(query)

	if err == nil {
		item.Attributes.Set("host", host)
		item.Attributes.Set("port", port)

		var linkType string

		// Add linked items depending on whether the host is an IP
		if net.ParseIP(host) == nil {
			linkType = "dns"
		} else {
			linkType = "ip"
		}

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    linkType,
			Method:  sdp.RequestMethod_GET,
			Query:   host,
			Context: itemContext,
		})
	} else {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	return &item, nil
}

// Find Is not implemented for HTTP as this would require scanning many
// endpoints or something, doesn't really make sense
func (s *SocketSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *SocketSource) Weight() int {
	return 100
}
