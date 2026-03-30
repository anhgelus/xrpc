package xrpc

import (
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
