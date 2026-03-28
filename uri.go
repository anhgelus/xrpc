package xrpc

import (
	"context"
	"errors"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

var ErrIncompleteURI = errors.New("incomplete URI")

// FetchURI is like [Client.FetchRawURI], but when the uri is determined.
func (c *BaseClient) FetchURI(ctx context.Context, uri atproto.URI) (*Union, error) {
	if uri.Collection() == nil || uri.RecordKey() == nil {
		return nil, ErrIncompleteURI
	}
	return rawGetRecord(ctx, c, uri.Authority(), uri.Collection(), *uri.RecordKey(), "")
}
