package internet

import (
	"context"
	"fmt"
	"net"

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
	Client  *rdap.Client
	Cache   *sdpcache.Cache
	IPCache *IPCache[*rdap.IPNetwork]
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
	hit, _, items, sdpErr := s.Cache.Lookup(ctx, s.Name(), sdp.QueryMethod_GET, scope, s.Type(), query, ignoreCache)

	if sdpErr != nil {
		return nil, sdpErr
	}

	if hit {
		if len(items) > 0 {
			return items[0], nil
		}
	}
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
	hit, ck, items, sdpErr := s.Cache.Lookup(ctx, s.Name(), sdp.QueryMethod_SEARCH, scope, s.Type(), query, ignoreCache)

	if sdpErr != nil {
		return nil, sdpErr
	}
	if hit {
		return items, nil
	}

	// Second layer of caching means that we cn look up an IP, and if there is
	// anything in the cache that covers a range that IP is in, it will hit
	// the cache
	var ipNetwork *rdap.IPNetwork

	// See which type of argument we have and parse it
	if ip := net.ParseIP(query); ip != nil {
		// Check if the IP is in the cache
		ipNetwork, hit = s.IPCache.SearchIP(ip)
	} else if _, network, err := net.ParseCIDR(query); err == nil {
		// Check if the CIDR is in the cache
		ipNetwork, hit = s.IPCache.SearchCIDR(network)
	} else {
		return nil, fmt.Errorf("Invalid IP or CIDR: %v", query)
	}

	if !hit {
		// If we didn't hit the cache, then actually execute the query
		request := &rdap.Request{
			Type:  rdap.IPRequest,
			Query: query,
		}
		request = request.WithContext(ctx)

		response, err := s.Client.Do(request)

		if err != nil {
			err = wrapRdapError(err)

			s.Cache.StoreError(err, CacheDuration, ck)

			return nil, err
		}

		if response.Object == nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				Scope:       scope,
				ErrorString: fmt.Sprintf("No IP Network found for %s", query),
				SourceName:  s.Name(),
			}
		}

		var ok bool

		ipNetwork, ok = response.Object.(*rdap.IPNetwork)

		if !ok {
			return nil, fmt.Errorf("Expected IPNetwork, got %T", response.Object)
		}

		// Calculate the CIDR for this network
		network, err := calculateNetwork(ipNetwork.StartAddress, ipNetwork.EndAddress)

		if err != nil {
			return nil, err
		}

		// Cache this network
		s.IPCache.Store(network, ipNetwork, CacheDuration)
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

	s.Cache.StoreItem(item, CacheDuration, ck)

	return []*sdp.Item{item}, nil
}

// Calculates the network (like a CIDR) from a given start and end IP
func calculateNetwork(startIP, endIP string) (*net.IPNet, error) {
	// Parse start and end IP addresses
	start := net.ParseIP(startIP)
	if start == nil {
		return nil, fmt.Errorf("Invalid start IP address: %s", startIP)
	}

	end := net.ParseIP(endIP)
	if end == nil {
		return nil, fmt.Errorf("Invalid end IP address: %s", endIP)
	}

	// Calculate the CIDR prefix length
	var prefixLen int
	for i := 0; i < len(start); i++ {
		startByte := start[i]
		endByte := end[i]

		if startByte != endByte {
			// Find the differing bit position
			diffBit := startByte ^ endByte

			// Count the number of consecutive zero bits in the differing byte
			for j := 7; j >= 0; j-- {
				if (diffBit & (1 << uint(j))) != 0 {
					break
				}
				prefixLen++
			}
			break
		}

		prefixLen += 8
	}

	mask := net.CIDRMask(int(prefixLen), 128)

	// Calculate the network address
	network := net.IPNet{
		IP:   start,
		Mask: mask,
	}

	return &network, nil
}
