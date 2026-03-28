package xrpc

import (
	"context"
	"errors"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

var ErrIncompleteURI = errors.New("incomplete URI")

func (c *BaseClient) FetchRawURI(ctx context.Context, uri atproto.RawURI) (*Union, error) {
	if uri.IsDID() {
		did, err := uri.DID()
		if err != nil {
			return nil, err
		}
		return FetchURI(ctx, c, did)
	} else if uri.IsHandle() {
		h, err := uri.Handle()
		if err != nil {
			return nil, err
		}
		return FetchURI(ctx, c, h)
	}
	panic("unsupported authority")
}

// FetchURI is like [Client.FetchRawURI], but when the uri is determined.
func FetchURI[A atproto.Authority](ctx context.Context, client Client, uri atproto.URI[A]) (*Union, error) {
	if uri.Collection() == nil || uri.RecordKey() == nil {
		return nil, ErrIncompleteURI
	}
	return rawGetRecord(ctx, client, uri.Authority(), uri.Collection(), *uri.RecordKey(), "")
}
