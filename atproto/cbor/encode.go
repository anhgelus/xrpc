package cbor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
)

var (
	ErrCannotEncodeUint  = errors.New("cannot encode uint (too big)")
	ErrCannotEncodeFloat = errors.New("cannot encode float")
	ErrUnsupportedType   = errors.New("unsupported type")
)

type majorType byte

const (
	unsignedInt majorType = 0b000
	negativeInt majorType = 0b001
	byteString  majorType = 0b010
	textString  majorType = 0b011
	array       majorType = 0b100
	mapT        majorType = 0b101
	// tag is not supported by atproto
	simpleValues majorType = 0b111
)

type additionalType byte

const (
	nextUint8        additionalType = 0b11000
	nextUint16       additionalType = 0b11001
	nextUint32       additionalType = 0b11010
	nextUint64       additionalType = 0b11011
	indefiniteLength additionalType = 0b11111
)

func createType(major majorType, additional additionalType) byte {
	if additional >= 1<<5 {
		panic("overflow: additional is too big")
	}
	return byte(major<<5) + byte(additional)
}

func Marshal(v any) ([]byte, error) {
	ref := reflect.ValueOf(v)
	switch ref.Kind() {
	case reflect.Int:
		val := ref.Int()
		if ref.Int() < 0 {
			val = (val * -1) - 1
			return marshalRawInt(negativeInt, uint64(val)), nil
		}
		ref = reflect.ValueOf(uint64(val))
		fallthrough
	case reflect.Uint:
		return marshalRawInt(unsignedInt, ref.Uint()), nil
	case reflect.Bool:
		var b additionalType = 20
		if ref.Bool() {
			b += 1
		}
		return []byte{createType(simpleValues, b)}, nil
	case reflect.Float32, reflect.Float64:
		return nil, ErrCannotEncodeFloat
	case reflect.String:
		return marshalBytes(textString, []byte(ref.String()))
	case reflect.Array, reflect.Slice:
		if val, ok := v.([]byte); ok {
			return marshalBytes(byteString, val)
		}
		ln := ref.Len()
		b := marshalRawInt(array, uint64(ln))
		for i := range ln {
			val, err := Marshal(ref.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			b = append(b, val...)
		}
	case reflect.Pointer:
		if ref.IsNil() {
			return []byte{createType(simpleValues, 22)}, nil
		}
	}
	return nil, fmt.Errorf("%w: %T", ErrUnsupportedType, v)
}

func marshalBytes(t majorType, sl []byte) ([]byte, error) {
	b := marshalRawInt(t, uint64(len(sl)))
	return append(b, sl...), nil
}

func marshalRawInt(t majorType, i uint64) []byte {
	if i <= 23 {
		return []byte{createType(t, additionalType(i))}
	}
	if i <= math.MaxUint8 {
		return []byte{createType(t, nextUint8), byte(i)}
	}
	b := make([]byte, 1)
	if i <= math.MaxUint16 {
		b[0] = createType(t, nextUint16)
		return binary.BigEndian.AppendUint16(b, uint16(i))
	}
	if i <= math.MaxUint32 {
		b[0] = createType(t, nextUint32)
		return binary.BigEndian.AppendUint32(b, uint32(i))
	}
	b[0] = createType(t, nextUint64)
	return binary.BigEndian.AppendUint64(b, i)
}
