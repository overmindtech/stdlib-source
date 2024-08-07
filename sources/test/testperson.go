package test

import (
	"context"

	"github.com/overmindtech/sdp-go"
)

// TestPersonSource A source of `person` items for automated tests.
type TestPersonSource struct{}

// Type is the type of items that this returns
func (s *TestPersonSource) Type() string {
	return "test-person"
}

// Name Returns the name of the backend
func (s *TestPersonSource) Name() string {
	return "stdlib-test-person"
}

// Weighting of duplicate sources
func (s *TestPersonSource) Weight() int {
	return 100
}

// List of scopes that this source is capable of find items for
func (s *TestPersonSource) Scopes() []string {
	return []string{
		"test",
	}
}

func (s *TestPersonSource) Hidden() bool {
	return true
}

// Gets a single item. This expects a name
func (d *TestPersonSource) Get(ctx context.Context, scope string, query string, ignoreCache bool) (*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "test-dylan":
		return dylan(), nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}

func (d *TestPersonSource) List(ctx context.Context, scope string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	return []*sdp.Item{dylan()}, nil
}

func (d *TestPersonSource) Search(ctx context.Context, scope string, query string, ignoreCache bool) ([]*sdp.Item, error) {
	if scope != "test" {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: "test queries only supported in 'test' scope",
			Scope:       scope,
		}
	}

	switch query {
	case "", "*", "test-dylan":
		return []*sdp.Item{dylan()}, nil
	default:
		return nil, &sdp.QueryError{
			ErrorType: sdp.QueryError_NOTFOUND,
			Scope:     scope,
		}
	}
}
