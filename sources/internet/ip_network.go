package internet

import (
	"context"
	"fmt"

	"github.com/openrdap/rdap"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdpcache"
)

//go:generate docgen ../../doc
// +overmind:type rdap-ip-network
// +overmind:search Search for the most specific network that contains the specified IP or CIDR
// +overmind:description Returns information about an IP network using the RDAP
// protocol. Only the `SEARCH` method should be used for this source since it's
// not possible to list all IP networks, and they can't be queried by "handle"
// which is the unique attribute

type IPNetworkSource struct {
	Client *rdap.Client
	Cache  *sdpcache.Cache
}

// Type is the type of items that this returns
func (s *IPNetworkSource) Type() string {
	return "rdap-ip-network"
}

// Name Returns the name of the source
func (s *IPNetworkSource) Name() string {
	return "rdap"
}

// Weighting of duplicate sources
func (s *IPNetworkSource) Weight() int {
	return 100
}

func (s *IPNetworkSource) Scopes() []string {
	return []string{
		"global",
	}
}

func (s *IPNetworkSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	// This source doesn't technically support the GET method (since you can't
	// use the handle to query for an IP network)
	return nil, &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: fmt.Sprintf("IP networks can't be queried by handle, use the SEARCH method instead"),
	}
}

func (s *IPNetworkSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	return nil, &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: fmt.Sprintf("IP networks cannot be listed, use the SEARCH method instead"),
	}
}

// Search for the most specific network that contains the specified IP or CIDR
func (s *IPNetworkSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	request := rdap.Request{
		Type:  rdap.IPRequest,
		Query: query,
	}

	response, err := s.Client.Do(&request)

	if err != nil {
		return nil, wrapRdapError(err)
	}

	if response.Object == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			Scope:       scope,
			ErrorString: fmt.Sprintf("No IP Network found for %s", query),
			SourceName:  s.Name(),
		}
	}

	ipNetwork, ok := response.Object.(*rdap.IPNetwork)

	if !ok {
		return nil, fmt.Errorf("Expected IPNetwork, got %T", response.Object)
	}

	attributes, err := sdp.ToAttributesSorted(map[string]interface{}{
		"conformance":     ipNetwork.Conformance,
		"country":         ipNetwork.Country,
		"endAddress":      ipNetwork.EndAddress,
		"events":          ipNetwork.Events,
		"handle":          ipNetwork.Handle,
		"ipVersion":       ipNetwork.IPVersion,
		"links":           ipNetwork.Links,
		"name":            ipNetwork.Name,
		"notices":         ipNetwork.Notices,
		"objectClassName": ipNetwork.ObjectClassName,
		"parentHandle":    ipNetwork.ParentHandle,
		"port43":          ipNetwork.Port43,
		"remarks":         ipNetwork.Remarks,
		"startAddress":    ipNetwork.StartAddress,
		"status":          ipNetwork.Status,
		"type":            ipNetwork.Type,
	})

	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            s.Type(),
		UniqueAttribute: "handle",
		Attributes:      attributes,
		Scope:           scope,
	}

	// Loop over the entities and create linkedin item queries
	item.LinkedItemQueries = extractEntityLinks(ipNetwork.Entities)

	return []*sdp.Item{item}, nil
}
