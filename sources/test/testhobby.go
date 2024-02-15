package test

import (
	"context"

	"github.com/overmindtech/sdp-go"
)

//go:generate docgen ../../doc
// +overmind:type test-hobby
// +overmind:get Returns pre-canned items for automated tests.
// +overmind:list Returns pre-canned items for automated tests.
// +overmind:search Returns pre-canned items for automated tests.

// +overmind:description This source reliably returns pre-canned items for automated tests.

// TestHobbySource A source of `hobby` items for automated tests.
type TestHobbySource struct{}

// Type is the type of items that this returns
func (s *TestHobbySource) Type() string {
	return "test-hobby"
}

// Name Returns the name of the backend
func (s *TestHobbySource) Name() string {
	return "stdlib-test-hobby"
}

// Weighting of duplicate sources
func (s *TestHobbySource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *TestHobbySource) Scopes() []string {
	return []string{
		"test",
	}
}

func (s *TestHobbySource) Hidden() bool {
	return true
}

// Gets a single item. This expects a name
func (d *TestHobbySource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "test-motorcycling":
		return motorcycling(), nil
	case "test-knitting":
		return knitting(), nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}

func (d *TestHobbySource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	return []*sdp.Item{motorcycling(), knitting()}, nil
}

func (d *TestHobbySource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "", "*", "test-motorcycling":
		return []*sdp.Item{motorcycling()}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}
