package cbor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrUnmarshalRequiresPointer = errors.New("unmarshal requires a pointer")
	ErrUnsupportedHead          = errors.New("unsupported head")
	ErrInvalidType              = errors.New("invalid type")
	ErrNotCBOR                  = errors.New("not a CBOR")
)

type byteReader struct {
	b []byte
	i uint
}

func (r *byteReader) Next() byte {
	b := r.b[r.i]
	r.i++
	return b
}

func (r *byteReader) More(i uint) []byte {
	b := r.b[r.i : r.i+i]
	r.i += i
	return b
}

func (r *byteReader) Peek() byte {
	return r.b[r.i]
}

func (r *byteReader) Drain() []byte {
	return r.b[r.i:]
}

type Unmarshaler interface {
	UnmarshalCBOR(b []byte) ([]byte, error)
}

func extractHead(r *byteReader) (majorType, additionalInformation) {
	b := r.Next()
	m := b & 0b111_00000
	a := b & 0b000_11111
	return majorType(m >> 5), additionalInformation(a)
}

func Unmarshal(b []byte, v any) (rest []byte, err error) {
	ref := reflect.ValueOf(v)
	if ref.Kind() != reflect.Pointer || ref.IsNil() {
		return nil, fmt.Errorf("%w: given %T", ErrUnmarshalRequiresPointer, v)
	}
	switch cv := v.(type) {
	case Unmarshaler:
		return cv.UnmarshalCBOR(b)
	}
	r := &byteReader{b: b}
	m, a := extractHead(r)

	var val any
	defer func() {
		v := recover()
		if v == nil {
			return
		}
		var ok bool
		err, ok = v.(error)
		if !ok {
			panic(err)
		}
		err = fmt.Errorf("%w: %w, for data [%x]", ErrNotCBOR, err, b)
	}()
	switch m {
	case unsignedInt:
		var t uint64
		t, err = unmarshalRawUint(a, r)
		if err != nil {
			return nil, err
		}
		val, err = unmarshalUint(ref.Elem().Kind(), t)
	case negativeInt:
		var t uint64
		t, err = unmarshalRawUint(a, r)
		if err != nil {
			return nil, err
		}
		val, err = unmarshalInt(ref.Elem().Kind(), int64(t)*-1-1)
	case byteString:
	case textString:
	case array:
	case mapT:
	case tag:
	case simpleValues:
	}
	if err != nil {
		return nil, err
	}
	ref.Elem().Set(reflect.ValueOf(val))
	return r.Drain(), nil
}

func unmarshalRawUint(a additionalInformation, r *byteReader) (uint64, error) {
	if a <= 23 {
		return uint64(a), nil
	}
	switch a {
	case nextUint8:
		return uint64(r.Next()), nil
	case nextUint16:
		return uint64(binary.BigEndian.Uint16(r.More(2))), nil
	case nextUint32:
		return uint64(binary.BigEndian.Uint32(r.More(4))), nil
	case nextUint64:
		return uint64(binary.BigEndian.Uint64(r.More(8))), nil
	default:
		return 0, fmt.Errorf("%w: %x", ErrUnsupportedHead, a)
	}
}

func unmarshalUint(kind reflect.Kind, val uint64) (any, error) {
	switch kind {
	case reflect.Uint8:
		return uint8(val), nil
	case reflect.Uint16:
		return uint16(val), nil
	case reflect.Uint32:
		return uint32(val), nil
	case reflect.Uint64:
		return uint64(val), nil
	case reflect.Uint:
		return uint(val), nil
	default:
		return unmarshalInt(kind, int64(val))
	}
}
func unmarshalInt(kind reflect.Kind, val int64) (any, error) {
	switch kind {
	case reflect.Int8:
		return int8(val), nil
	case reflect.Int16:
		return int16(val), nil
	case reflect.Int32:
		return int32(val), nil
	case reflect.Int64:
		return int64(val), nil
	case reflect.Int:
		return int(val), nil
	}
	return nil, fmt.Errorf("%w: %v is not an (u)int", ErrInvalidType, kind)
}
