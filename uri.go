package xrpc

import (
	"context"
	"errors"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

var ErrIncompleteURI = errors.New("incomplete URI")

// FetchURI is like [Client.FetchRawURI], but when the uri is determined.
func (c *BaseClient) FetchURI(ctx context.Context, uri atproto.URI) (RecordStored[*Union], error) {
	var v RecordStored[*Union]
	if uri.Collection() == nil || uri.RecordKey() == nil {
		return v, ErrIncompleteURI
	}
	return rawGetRecord(ctx, c, uri.Authority(), uri.Collection(), *uri.RecordKey(), nil)
}
