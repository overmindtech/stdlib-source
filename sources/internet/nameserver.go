package internet

import (
	"context"
	"fmt"
	"strings"

	"github.com/openrdap/rdap"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdpcache"
)

//go:generate docgen ../../doc
// +overmind:type rdap-nameserver
// +overmind:search Search for the RDAP entry for a nameserver by its full URL e.g. "https://rdap.verisign.com/com/v1/nameserver/NS4.GOOGLE.COM"
// +overmind:description Returns details from RDAP about nameservers

type NameserverSource struct {
	ClientFac func() *rdap.Client
	Cache     *sdpcache.Cache
}

// Type is the type of items that this returns
func (s *NameserverSource) Type() string {
	return "rdap-nameserver"
}

// Name Returns the name of the source
func (s *NameserverSource) Name() string {
	return "rdap"
}

// Weighting of duplicate sources
func (s *NameserverSource) Weight() int {
	return 100
}

func (s *NameserverSource) Scopes() []string {
	return []string{
		"global",
	}
}

func (s *NameserverSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	// Check the cache for GET requests, if we don't hit the cache then there is
	// nothing we can do though
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
		ErrorString: fmt.Sprintf("Nameservers can't be queried by handle, use the SEARCH method instead"),
	}
}

func (s *NameserverSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	return nil, &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: fmt.Sprintf("Nameservers cannot be listed, use the SEARCH method instead"),
	}
}

// Search for the nameserver using the full RDAP URL. This is required since
// nameserver queries are not capable of being bootstrapped and we need to know
// which nameserver to query from the beginning. Fortunately domain queries can
// be bootstrapped, so we can use the domain query to find the nameserver in the
// link
func (s *NameserverSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	hit, ck, items, sdpErr := s.Cache.Lookup(ctx, s.Name(), sdp.QueryMethod_SEARCH, scope, s.Type(), query, ignoreCache)

	if sdpErr != nil {
		return nil, sdpErr
	}
	if hit {
		return items, nil
	}

	parsed, err := parseRdapUrl(query)

	if err != nil {
		return nil, err
	}

	request := &rdap.Request{
		Type:   rdap.NameserverRequest,
		Query:  parsed.Query,
		Server: parsed.ServerRoot,
	}
	request.WithContext(ctx)

	response, err := s.ClientFac().Do(request)

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

	nameserver, ok := response.Object.(*rdap.Nameserver)

	if !ok {
		return nil, fmt.Errorf("Expected Nameserver, got %T", response.Object)
	}

	attributes, err := sdp.ToAttributesCustom(map[string]interface{}{
		"conformance":     nameserver.Conformance,
		"objectClassName": nameserver.ObjectClassName,
		"notices":         nameserver.Notices,
		"handle":          nameserver.Handle,
		"ldhName":         nameserver.LDHName,
		"unicodeName":     nameserver.UnicodeName,
		"ipAddresses":     nameserver.IPAddresses,
		"status":          nameserver.Status,
		"remarks":         nameserver.Remarks,
		"links":           nameserver.Links,
		"port43":          nameserver.Port43,
		"events":          nameserver.Events,
	}, true, RDAPTransforms)

	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            s.Type(),
		UniqueAttribute: "ldhName",
		Attributes:      attributes,
		Scope:           scope,
	}

	// Link entities
	// +overmind:link rdap-entity
	item.LinkedItemQueries = append(item.LinkedItemQueries, extractEntityLinks(nameserver.Entities)...)

	// Nameservers are resolvable in DNS too
	// +overmind:link dns
	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   "dns",
			Method: sdp.QueryMethod_SEARCH,
			Query:  strings.ToLower(nameserver.LDHName),
			Scope:  "global",
		},
		BlastPropagation: &sdp.BlastPropagation{
			// These represent the same thing so linked them both ways
			In:  true,
			Out: true,
		},
	})

	// Link IP addresses
	if nameserver.IPAddresses != nil {
		allIPs := append(nameserver.IPAddresses.V4, nameserver.IPAddresses.V6...)

		for _, ip := range allIPs {
			// +overmind:link ip
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  ip,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// IPs are always linked
					In:  true,
					Out: true,
				},
			})
		}
	}

	s.Cache.StoreItem(item, CacheDuration, ck)

	return []*sdp.Item{item}, nil
}
