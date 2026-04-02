package xrpc

import (
	"context"
	"encoding/json"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

var CollectionBlob = &atproto.NSID{Name: "blob"}

// Blob represents an ATProto blob.
type Blob struct {
	CID      *atproto.CIDLink `json:"ref"`
	MimeType string           `json:"mimeType"`
	Size     uint             `json:"size"`
}

func (b *Blob) Collection() *atproto.NSID {
	return CollectionBlob
}

// UploadBlob to be referenced from a [Record] for the authentificated [Client].
// The blob will be deleted if it is not referenced within a time window (eg, minutes).
// Blob restrictions (mimetype, size, etc) are enforced when the reference is created.
func UploadBlob(ctx context.Context, client Client, contentType string, blob []byte) (*Blob, error) {
	req := client.NewRequest().Endpoint(collection.Name("UploadBlob").Build())
	b, err := client.Procedure(ctx, req, RawBodyRequest{blob, contentType})
	if err != nil {
		return nil, err
	}
	var v struct {
		Blob *Blob `json:"blob"`
	}
	return v.Blob, json.Unmarshal(b, &v)
}
