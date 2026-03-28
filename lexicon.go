package xrpc

import (
	"encoding/json"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// Record represents an ATProto record.
type Record interface {
	// Type returns the [atproto.NSID] of the lexicon behind the [Record].
	// Must be stateless.
	Type() *atproto.NSID
}

// Union represents an ATProto *open* union.
//
// See [UnionAs] to convert an [Union] into a [Record].
type Union struct {
	tpe *atproto.NSID
	Raw []byte
}

func (u *Union) Type() *atproto.NSID {
	return u.tpe
}

func (u *Union) UnmarshalJSON(b []byte) error {
	var v struct {
		Type *atproto.NSID `json:"string"`
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	u.tpe = v.Type
	u.Raw = b
	return nil
}

// As converts an [Union] into a [Record].
//
// Returns false if it cannot convert.
func (u *Union) As(rec Record) bool {
	if rec.Type() != u.Type() {
		return false
	}
	err := json.Unmarshal(u.Raw, rec)
	return err == nil
}
