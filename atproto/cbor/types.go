package cbor

import (
	"fmt"
	"reflect"
	"strings"
)

type majorType byte

const (
	unsignedInt  majorType = 0b000
	negativeInt  majorType = 0b001
	byteString   majorType = 0b010
	textString   majorType = 0b011
	array        majorType = 0b100
	mapT         majorType = 0b101
	tag          majorType = 0b110
	simpleValues majorType = 0b111
)

type additionalInformation byte

const (
	nextUint8  additionalInformation = 0b11000
	nextUint16 additionalInformation = 0b11001
	nextUint32 additionalInformation = 0b11010
	nextUint64 additionalInformation = 0b11011
	//indefiniteLength additionalInformation = 0b11111
)

// Tag implements a CBOR Tag.
type Tag interface {
	// Tag returns the uint of the [Tag].
	Tag() uint64
	// Content returns the content of the [Tag].
	Content() any
	// Unmarshal the content of the [Tag] with the given number and the remaining bytes.
	// See [Unmarshaler].
	UnmarshalCBOR(uint64, *ByteReader) ([]byte, error)
}

func marshalTag(t Tag) ([]byte, error) {
	in, err := Marshal(t.Content())
	if err != nil {
		return nil, err
	}
	b := marshalRawInt(tag, t.Tag())
	return append(b, in...), nil
}

func unmarshalTag(b []byte, t Tag) ([]byte, error) {
	r := &ByteReader{Bytes: b}
	m, a := extractHead(r)
	if m != tag {
		return nil, fmt.Errorf("%w: %v is not a tag, for data [% x]", ErrInvalidType, m, b)
	}
	v, err := unmarshalRawUint(a, r)
	if err != nil {
		return nil, err
	}
	return t.UnmarshalCBOR(v, r)
}

type options struct {
	name      string
	omitempty bool
	toString  bool
}

func optionsOf(field reflect.StructField) options {
	var opts options
	if tag, ok := field.Tag.Lookup("cbor"); ok {
		opts = parseTag(tag)
	} else if tag, ok := field.Tag.Lookup("json"); ok {
		opts = parseTag(tag)
	}
	if opts.name == "" {
		opts.name = field.Name
	}
	return opts
}

func parseTag(tag string) options {
	parts := strings.Split(tag, ",")
	opts := options{name: parts[0]}
	if len(parts) == 1 {
		return opts
	}
	for _, p := range parts[1:] {
		switch p {
		case "omitempty":
			opts.omitempty = true
		case "string":
			opts.toString = true
		default:
			panic("unsupported options: " + p)
		}
	}
	return opts
}

// ByteReader is a helper to read bytes.
type ByteReader struct {
	Bytes []byte
	I     uint
}

// Next returns the next byte.
// It doesn't perform sanity checks.
func (r *ByteReader) Next() byte {
	b := r.Bytes[r.I]
	r.I++
	return b
}

// Consume the current byte without returning it.
// Returns true if everything was read.
func (r *ByteReader) Consume() bool {
	r.I++
	return int(r.I) < len(r.Bytes)
}

// More performs multiple calls to [Next].
func (r *ByteReader) More(i uint) []byte {
	b := r.Bytes[r.I : r.I+i]
	r.I += i
	return b
}

// Peek return the next byte without consuming it.
func (r *ByteReader) Peek() byte {
	return r.Bytes[r.I]
}

// Drain the remaining bytes and return them.
func (r *ByteReader) Drain() []byte {
	if int(r.I) >= len(r.Bytes) {
		return nil
	}
	return r.Bytes[r.I:]
}
