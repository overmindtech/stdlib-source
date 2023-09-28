package test

import (
	"context"

	"github.com/overmindtech/sdp-go"
)

//go:generate docgen ../../doc
// +overmind:type test-dog
// +overmind:get Returns pre-canned items for automated tests.
// +overmind:list Returns pre-canned items for automated tests.
// +overmind:search Returns pre-canned items for automated tests.

// +overmind:description This source reliably returns pre-canned items for automated tests.

// TestDogSource A source of `dog` items for automated tests.
type TestDogSource struct{}

// Type is the type of items that this returns
func (s *TestDogSource) Type() string {
	return "test-dog"
}

// Name Returns the name of the backend
func (s *TestDogSource) Name() string {
	return "stdlib-test-dog"
}

// Weighting of duplicate sources
func (s *TestDogSource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *TestDogSource) Scopes() []string {
	return []string{
		"test",
	}
}

func (s *TestDogSource) Hidden() bool {
	return true
}

// Gets a single item. This expects a name
func (d *TestDogSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "test-manny":
		return manny(), nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}

func (d *TestDogSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	return []*sdp.Item{manny()}, nil
}

func (d *TestDogSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "", "*", "test-manny":
		return []*sdp.Item{manny()}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}
