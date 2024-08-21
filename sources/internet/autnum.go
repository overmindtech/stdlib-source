package internet

import (
	"context"
	"fmt"
	"strings"

	"github.com/openrdap/rdap"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdpcache"
)

//go:generate docgen ../../docs-data
// +overmind:type rdap-asn
// +overmind:descriptiveType Autonomous System Number (ASN)
// +overmind:get Get an ASN by handle i.e. "AS15169"
// +overmind:description Uses RDAP to get the details of an Autonomous System
// Number (ASN) by number

type ASNSource struct {
	ClientFac func() *rdap.Client
	Cache     *sdpcache.Cache
}

// Type is the type of items that this returns
func (s *ASNSource) Type() string {
	return "rdap-asn"
}

// Name Returns the name of the backend
func (s *ASNSource) Name() string {
	return "rdap"
}

// Weighting of duplicate sources
func (s *ASNSource) Weight() int {
	return 100
}

func (s *ASNSource) Scopes() []string {
	return []string{
		"global",
	}
}

func (s *ASNSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	hit, ck, items, sdpErr := s.Cache.Lookup(ctx, s.Name(), sdp.QueryMethod_GET, scope, s.Type(), query, ignoreCache)

	if sdpErr != nil {
		return nil, sdpErr
	}

	if hit {
		if len(items) > 0 {
			return items[0], nil
		}
	}

	// Strip the AS prefix
	query = strings.TrimPrefix(query, "AS")

	request := &rdap.Request{
		Type:  rdap.AutnumRequest,
		Query: query,
	}
	request = request.WithContext(ctx)

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
			ErrorString: "No ASN found",
		}
	}

	asn, ok := response.Object.(*rdap.Autnum)

	if !ok {
		return nil, fmt.Errorf("Unexpected response type: %T", response.Object)
	}

	attributes, err := sdp.ToAttributesCustom(map[string]interface{}{
		"conformance":     asn.Conformance,
		"objectClassName": asn.ObjectClassName,
		"notices":         asn.Notices,
		"handle":          asn.Handle,
		"startAutnum":     asn.StartAutnum,
		"endAutnum":       asn.EndAutnum,
		"ipVersion":       asn.IPVersion,
		"name":            asn.Name,
		"type":            asn.Type,
		"status":          asn.Status,
		"country":         asn.Country,
		"remarks":         asn.Remarks,
		"links":           asn.Links,
		"port43":          asn.Port43,
		"events":          asn.Events,
	}, true, RDAPTransforms)

	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            s.Type(),
		UniqueAttribute: "handle",
		Attributes:      attributes,
		Scope:           scope,
	}

	// Link the entities
	// +overmind:link rdap-entity
	item.LinkedItemQueries = extractEntityLinks(asn.Entities)

	s.Cache.StoreItem(item, CacheDuration, ck)

	return item, nil
}

func (s *ASNSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	return nil, &sdp.QueryError{
		ErrorType:   sdp.QueryError_NOTFOUND,
		Scope:       scope,
		ErrorString: "ASNs cannot be listed, use the GET method instead",
	}
}
