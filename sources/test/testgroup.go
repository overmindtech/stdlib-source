package test

import (
	"context"

	"github.com/overmindtech/sdp-go"
)

//go:generate docgen ../../doc
// +overmind:type test-group
// +overmind:get Returns pre-canned items for automated tests.
// +overmind:list Returns pre-canned items for automated tests.
// +overmind:search Returns pre-canned items for automated tests.

// +overmind:description This source reliably returns pre-canned items for automated tests.

// TestGroupSource A source of `group` items for automated tests.
type TestGroupSource struct{}

// Type is the type of items that this returns
func (s *TestGroupSource) Type() string {
	return "test-group"
}

// Name Returns the name of the backend
func (s *TestGroupSource) Name() string {
	return "stdlib-test-group"
}

// Weighting of duplicate sources
func (s *TestGroupSource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *TestGroupSource) Scopes() []string {
	return []string{
		"test",
	}
}

func (s *TestGroupSource) Hidden() bool {
	return true
}

// Gets a single item. This expects a name
func (d *TestGroupSource) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "test-admins":
		return admins(), nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}

func (d *TestGroupSource) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	return []*sdp.Item{admins()}, nil
}

func (d *TestGroupSource) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "", "*", "test-admins":
		return []*sdp.Item{admins()}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}
