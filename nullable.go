package xrpc

import (
	"encoding/json"

	"anhgelus.world/xrpc/atproto/cbor"
)

// Nullable represents a type that can be null and absent.
// It is intended to be used with omitempty.
//
// See [MapNullable].
type Nullable[T any] struct {
	data *T
}

func (n *Nullable[T]) MarshalMap() (any, error) {
	if !n.Present() {
		return nil, nil
	}
	return MarshalToMap(n.data)
}

func (n *Nullable[T]) MarshalCBOR() ([]byte, error) {
	if !n.Present() {
		return cbor.Marshal(nil)
	}
	return cbor.Marshal(n.data)
}

func (n *Nullable[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, n.data)
}

func (n *Nullable[T]) UnmarshalCBOR(b []byte) ([]byte, error) {
	return cbor.Unmarshal(b, n.data)
}

// Get returns the content contained.
// Value is nil if it is null.
func (n *Nullable[T]) Get() *T {
	return n.data
}

// Present returns true if the value is present.
func (n *Nullable[T]) Present() bool {
	return n.data != nil
}

// Set the value in Nullable.
func (n *Nullable[T]) Set(v *T) {
	n.data = v
}

// MapNullable is the equivalent of a functional map applied to [Nullable].
func MapNullable[A, B any](n *Nullable[A], fn func(*A) *B) *Nullable[B] {
	if n == nil {
		return nil
	}
	if !n.Present() {
		return &Nullable[B]{}
	}
	return &Nullable[B]{fn(n.data)}
}
