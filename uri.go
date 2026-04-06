package xrpc

import (
	"context"
	"encoding/json"
	"errors"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// ErrIncompleteURI is returned when the [atproto.URI] does not have enough information to be used.
var ErrIncompleteURI = errors.New("incomplete URI")

func (c *BaseClient) FetchURI(ctx context.Context, uri atproto.URI) (RecordStored[*Union], error) {
	var v RecordStored[*Union]
	if uri.Collection() == nil || uri.RecordKey() == nil {
		return v, ErrIncompleteURI
	}
	b, err := rawGetRecord(ctx, c, uri.Authority(), uri.Collection(), *uri.RecordKey(), nil)
	if err != nil {
		return v, err
	}
	return v, json.Unmarshal(b, &v)
}
