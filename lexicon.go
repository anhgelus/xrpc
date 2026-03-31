package xrpc

import (
	"context"
	"encoding/json"
	"errors"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// Record represents an ATProto record.
//
// If it implements [MapMarshaler], it must returns a map[string]any.
type Record interface {
	// Collection returns the [atproto.NSID] of the lexicon behind the [Record].
	// Must be stateless.
	Collection() *atproto.NSID
}

var ErrNotRecord = errors.New("value got is not a record")

// Union represents an ATProto *open* union or an unknown type.
//
// See [Union.As] to convert an [Union] into a [Record].
// See [AsUnion] to convert a [Record] into an [Union].
type Union struct {
	tpe *atproto.NSID
	// Raw is set when the [Union] is unmarshaled.
	Raw []byte
	// Content is set if the [Union] is created from a [Record] with [AsUnion].
	Content Record
}

// AsUnion converts a [Record] to an [Union].
func AsUnion(rec Record) *Union {
	return &Union{tpe: rec.Collection(), Content: rec}
}

func (u *Union) Collection() *atproto.NSID {
	return u.tpe
}

func (u *Union) MarshalMap() (any, error) {
	if u.Content != nil {
		return Marshal(u.Content)
	}
	return u.Raw, nil
}

func (u *Union) UnmarshalJSON(b []byte) error {
	var v struct {
		Type *atproto.NSID `json:"$type"`
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	if v.Type == nil {
		return ErrNotRecord
	}
	u.tpe = v.Type
	u.Raw = b
	return nil
}

// As converts an [Union] into a [Record].
//
// Returns false if it cannot convert.
func (u *Union) As(rec Record) bool {
	if !u.Collection().Is(rec.Collection()) {
		return false
	}
	err := json.Unmarshal(u.Raw, rec)
	return err == nil
}

// Marshal a [Record] into a JSON.
func Marshal(a any) ([]byte, error) {
	v, err := MarshalToMap(a)
	if err != nil {
		return nil, err
	}
	if rec, ok := a.(Record); ok {
		v.(map[string]any)["$type"] = rec.Collection()
	}
	return json.Marshal(v)
}

var collection = atproto.NewNSIDBuilder("com.atproto.repo")

var CollectionStrongRef = collection.Name("strongRef").Build()

// StrongRef is an [atproto.RawURI] with a content-hash fingerprint.
//
// It doesn't implement [Record] because it is an object and not a record.
type StrongRef struct {
	URI atproto.RawURI `json:"uri"`
	CID string         `json:"cid"`
}

// GetRef returns an [Union] containing the [Record] pointed by the [StrongRef].
func (s *StrongRef) GetRef(ctx context.Context, client Client) (*Union, error) {
	uri, err := s.URI.URI(ctx, client.Directory())
	if err != nil {
		return nil, err
	}
	union, err := client.FetchURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return union.Value, nil
}

var CollectionCommitMeta = collection.Name("defs").Fragment("commitMeta").Build()

// CommitMeta contains information about a certain commit.
//
// It doesn't implement [Record] because it is an object and not a record.
type CommitMeta struct {
	CID *atproto.CIDAsString `json:"cid"`
	Rev atproto.TID          `json:"tid"`
}
