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
// +overmind:type rdap-domain
// +overmind:search Search for a domain record by the domain name e.g. "www.google.com"
// +overmind:description This source returns information about a domain using the
// RDAP protocol. The `SEARCH` method should be used for this source since it's
// not possible to list all domains, and they can't be queried by "handle" which
// is the unique attribute

type DomainSource struct {
	Client *rdap.Client
	Cache  *sdpcache.Cache
}

// Type is the type of items that this returns
func (s *DomainSource) Type() string {
	return "rdap-domain"
}

// Name Returns the name of the backend
func (s *DomainSource) Name() string {
	return "rdap"
}

// Weighting of duplicate sources
func (s *DomainSource) Weight() int {
	return 100
}

func (s *DomainSource) Scopes() []string {
	return []string{
		"global",
	}
}

func (s *DomainSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	return nil, &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: "Domains can't be queried by handle, use the SEARCH method instead",
	}
}

func (s *DomainSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	return nil, &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: "Domains listed, use the SEARCH method instead",
	}
}

// Search for the most specific domain that contains the specified domain. The
// input should be something like "www.google.com". This will first search for
// "www.google.com", then "google.com", then "com"
func (s *DomainSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	hit, ck, items, sdpErr := s.Cache.Lookup(ctx, s.Name(), sdp.QueryMethod_SEARCH, scope, s.Type(), query, ignoreCache)

	if sdpErr != nil {
		return nil, sdpErr
	}
	if hit {
		return items, nil
	}

	// Split the query into subdomains
	sections := strings.Split(query, ".")

	// Start by querying the whole domain, then go down from there, however
	// don't query for the top-level domain as it won't return anything useful
	for i := 0; i < len(sections)-1; i++ {
		domainName := strings.Join(sections[i:], ".")

		request := &rdap.Request{
			Type:  rdap.DomainRequest,
			Query: domainName,
		}

		response, err := s.Client.Do(request)

		if err != nil {
			// If there was an error, continue to the next domain
			continue
		}

		if response.Object == nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				Scope:       scope,
				ErrorString: "Empty domain response",
			}
		}

		domain, ok := response.Object.(*rdap.Domain)

		if !ok {
			return nil, fmt.Errorf("Unexpected response type %T", response.Object)
		}

		attributes, err := sdp.ToAttributesSorted(map[string]interface{}{
			"conformance":     domain.Conformance,
			"events":          domain.Events,
			"handle":          domain.Handle,
			"ldhName":         domain.LDHName,
			"links":           domain.Links,
			"notices":         domain.Notices,
			"objectClassName": domain.ObjectClassName,
			"port43":          domain.Port43,
			"publicIDs":       domain.PublicIDs,
			"remarks":         domain.Remarks,
			"secureDNS":       domain.SecureDNS,
			"status":          domain.Status,
			"unicodeName":     domain.UnicodeName,
			"variants":        domain.Variants,
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

		// Link to nameservers
		for _, nameServer := range domain.Nameservers {
			// Look through the HTTP responses until we find one
			var parsed *RDAPUrl
			for _, httpResponse := range response.HTTP {
				if httpResponse.URL != "" {
					parsed, err = parseRdapUrl(httpResponse.URL)

					if err == nil {
						break
					}
				}
			}

			// Reconstruct the required query URL
			if parsed != nil {
				newURL := parsed.ServerRoot.JoinPath("/nameserver/" + nameServer.LDHName)

				// +overmind:link rdap-nameserver
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "rdap-nameserver",
						Method: sdp.QueryMethod_SEARCH,
						Query:  newURL.String(),
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// A change in a name server could affect the domains
						In: true,
						// Domains won't affect the name server
						Out: false,
					},
				})
			}

		}

		// Link to entities
		// +overmind:link rdap-entity
		item.LinkedItemQueries = append(item.LinkedItemQueries, extractEntityLinks(domain.Entities)...)

		// Link to IP Network
		if network := domain.Network; network != nil {
			// +overmind:link rdap-ip-network
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "rdap-ip-network",
					Method: sdp.QueryMethod_SEARCH,
					Query:  network.StartAddress,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the network could affect the domain presumably
					In: true,
					// The domain won't affect the network
					Out: false,
				},
			})
		}

		if err != nil {
			return nil, err
		}

		s.Cache.StoreItem(item, CacheDuration, ck)

		return []*sdp.Item{item}, nil
	}

	err := &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: fmt.Sprintf("No domain found for %s", query),
	}

	s.Cache.StoreError(err, CacheDuration, ck)

	return nil, err
}
