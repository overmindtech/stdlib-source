package test

import (
	"context"

	"github.com/overmindtech/sdp-go"
)

// TestRegionSource A source of `region` items for automated tests.
type TestRegionSource struct{}

// Type is the type of items that this returns
func (s *TestRegionSource) Type() string {
	return "test-region"
}

// Name Returns the name of the backend
func (s *TestRegionSource) Name() string {
	return "stdlib-test-region"
}

// Weighting of duplicate sources
func (s *TestRegionSource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *TestRegionSource) Scopes() []string {
	return []string{
		"test",
	}
}

func (s *TestRegionSource) Hidden() bool {
	return true
}

// Gets a single item. This expects a name
func (d *TestRegionSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "test-gb":
		return gb(), nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}

func (d *TestRegionSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	return []*sdp.Item{gb()}, nil
}

func (d *TestRegionSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "", "*", "test-gb":
		return []*sdp.Item{gb()}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}
