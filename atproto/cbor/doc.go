// Package cbor implements the DRISL CBOR specified by AT Protocol.
// You can encode any data and decode CBOR.
// This package provides a reflection interface similar to Go's [encoding/json].
//
// [Marshal] is used to encode any structure into CBOR.
// You can implement [Marshaler] to have a custom behavior.
//
// [Unmarshal] is used to decode any CBOR into a structure.
// You can implement [Unmarshaler] to have a custom behavior.
//
// If uses the field's tag `cbor` to set properties like the field name.
// For example,
//
//	type Foo struct {
//		A bool `cbor:"a"`
//		B uint `cbor:"b"`
//	}
//
// will create a CBOR map with `a` and `b` as keys.
// You can use `omitempty` if it must omit the field if it's value is zero (see [reflect.Zero]).
// You can use `string` if it must be converted into a string while encoding.
//
// If there is no `cbor` tag, it uses the common `json` tag used by [encoding/json].
package cbor
