package xrpc

import (
	"encoding/json"
	"errors"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// Record represents an ATProto record.
type Record interface {
	// Type returns the [atproto.NSID] of the lexicon behind the [Record].
	// Must be stateless.
	Type() *atproto.NSID
}

var ErrNotRecord = errors.New("value got is not a record")

// Union represents an ATProto *open* union.
//
// See [Union.As] to convert an [Union] into a [Record].
// See [AsUnion] to convert a [Record] into an [Union].
type Union struct {
	tpe *atproto.NSID
	Raw []byte
}

// AsUnion converts a [Record] to an [Union].
//
// Returns an error if cannot marshal [Record].
func AsUnion(rec Record) (*Union, error) {
	union := &Union{tpe: rec.Type()}
	t, err := MarshalToMap(rec)
	if err != nil {
		return nil, err
	}
	union.Raw, err = json.Marshal(t)
	return union, err
}

func (u *Union) Type() *atproto.NSID {
	return u.tpe
}

func (u *Union) MarshalJSON() ([]byte, error) {
	return u.Raw, nil
}

func (u *Union) MarshalMap() (any, error) {
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
	if !u.Type().Is(rec.Type()) {
		return false
	}
	err := json.Unmarshal(u.Raw, rec)
	return err == nil
}

// Marshal a [Record] into a JSON.
func Marshal(rec Record) ([]byte, error) {
	v, err := MarshalToMap(rec)
	if err != nil {
		return nil, err
	}
	mp := v.(map[string]any)
	mp["$type"] = rec.Type()
	return json.Marshal(mp)
}
