package xrpc

import (
	"encoding/json"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

var CollectionBlob = &atproto.NSID{Name: "blob"}

// Blob represents an ATProto `blob` type.
type Blob struct {
	CID      string `json:"-"`
	MimeType string `json:"mimeType"`
	Size     uint   `json:"size"`
}

func (b *Blob) Type() *atproto.NSID {
	return CollectionBlob
}

func (b *Blob) MarshalMap() (any, error) {
	mp := make(map[string]any, 3)
	mp["mimeType"] = b.MimeType
	mp["size"] = b.Size
	mp["ref"] = map[string]any{"$link": b.CID}
	return mp, nil
}

func (b *Blob) UnmarshalJSON(data []byte) error {
	type t Blob
	var v struct {
		t
		Ref struct {
			Link string `json:"$link"`
		} `json:"ref"`
	}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	*b = Blob(v.t)
	b.CID = v.Ref.Link
	return nil
}
