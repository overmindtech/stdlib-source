package internet

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"

	"github.com/openrdap/rdap"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdpcache"
)

// Create sources from this package, these sources will share a cache, http
// client, and rdap client
func NewSources() []discovery.Source {
	// TODO: Change this to something from Otel
	httpClient := http.DefaultClient
	rdapClient := &rdap.Client{
		HTTP: httpClient,
	}
	cache := sdpcache.NewCache()

	return []discovery.Source{
		&IPNetworkSource{
			Client: rdapClient,
			Cache:  cache,
		},
	}
}

// Wraps an RDAP error in an SDP error, correctly checking for things like 404s
func wrapRdapError(err error) error {
	if err == nil {
		return nil
	}

	var rdapError *rdap.ClientError

	if ok := errors.As(err, &rdapError); ok {
		if rdapError.Type == rdap.ObjectDoesNotExist {
			return &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: err.Error(),
			}
		}
	}

	return err
}

// Extracts SDP queries from a list of entities
func extractEntityLinks(entities []rdap.Entity) []*sdp.LinkedItemQuery {
	queries := make([]*sdp.LinkedItemQuery, 0)

	for _, entity := range entities {
		var selfLink string

		// Loop over the links until you find the self link
		for _, link := range entity.Links {
			if link.Rel == "self" {
				selfLink = link.Href
				break
			}
		}

		if selfLink != "" {
			// +overmind:linked-item-query rdap-entity
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "rdap-entity",
					Method: sdp.QueryMethod_SEARCH,
					Query:  selfLink,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The Entity isn't a "real" component, so no matter what
					// changes it won't actually "affect" anything
					In:  false,
					Out: false,
				},
			})
		}
	}

	return queries
}

var rdapUrlRegex = regexp.MustCompile(`^(https?:\/\/.+)\/(ip|nameserver|entity|autnum|domain)\/([^\/]+)$`)

type RDAPUrl struct {
	// The path to the root where queries should be run i.e.
	// https://rdap.apnic.net
	ServerRoot *url.URL
	// The type of query to run i.e. ip, nameserver, entity, autnum, domain
	Type string
	// The query to run i.e. 1.1.1.1
	Query string
}

// Parses an RDAP URL and returns the important components
func parseRdapUrl(rdapUrl string) (*RDAPUrl, error) {
	matches := rdapUrlRegex.FindStringSubmatch(rdapUrl)

	if len(matches) != 4 {
		return nil, errors.New("Invalid RDAP URL")
	}

	serverRoot, err := url.Parse(matches[1])

	if err != nil {
		return nil, err
	}

	return &RDAPUrl{
		ServerRoot: serverRoot,
		Type:       matches[2],
		Query:      matches[3],
	}, nil
}
