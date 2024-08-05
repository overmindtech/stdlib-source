package test

import (
	"context"

	"github.com/overmindtech/sdp-go"
)

// TestLocationSource A source of `location` items for automated tests.
type TestLocationSource struct{}

// Type is the type of items that this returns
func (s *TestLocationSource) Type() string {
	return "test-location"
}

// Name Returns the name of the backend
func (s *TestLocationSource) Name() string {
	return "stdlib-test-location"
}

// Weighting of duplicate sources
func (s *TestLocationSource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *TestLocationSource) Scopes() []string {
	return []string{
		"test",
	}
}

func (s *TestLocationSource) Hidden() bool {
	return true
}

// Gets a single item. This expects a name
func (d *TestLocationSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "test-london":
		return london(), nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}

func (d *TestLocationSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	return []*sdp.Item{london()}, nil
}

func (d *TestLocationSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "", "*", "test-london":
		return []*sdp.Item{london()}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}
