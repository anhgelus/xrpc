package cbor

import (
	"cmp"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
)

var (
	ErrCannotEncodeFloat = errors.New("cannot encode float")
	ErrUnsupportedType   = errors.New("unsupported type")
	ErrInvalidMap        = errors.New("invalid map")
)

func createType(major majorType, additional additionalInformation) byte {
	if additional >= 1<<5 {
		panic("overflow: additional is too big")
	}
	return byte(major<<5) + byte(additional)
}

// Marshaler describes how the type is encoded into CBOR.
type Marshaler interface {
	MarshalCBOR() ([]byte, error)
}

// Marshal a value into CBOR.
func Marshal(v any) ([]byte, error) {
	if v != nil {
		switch cv := v.(type) {
		case Marshaler:
			return cv.MarshalCBOR()
		case Tag:
			return marshalTag(cv)
		}
	}
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
		var b additionalInformation = 20
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
		return b, nil
	case reflect.Pointer:
		if ref.IsNil() {
			return []byte{createType(simpleValues, 22)}, nil
		}
		elem := ref.Elem()
		if !elem.CanInterface() {
			return nil, nil
		}
		return Marshal(elem.Interface())
	case reflect.Struct:
		mp, err := structToMap(ref)
		if err != nil {
			return nil, err
		}
		ref = reflect.ValueOf(mp)
		fallthrough
	case reflect.Map:
		r := ref.MapRange()
		b := marshalRawInt(mapT, uint64(ref.Len()))
		mp := make([][2][]byte, ref.Len())
		for r.Next() {
			k := r.Key()
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("%w: keys must be strings", ErrInvalidMap)
			}
			in, err := marshalKeyVal(k.String(), r.Value().Interface())
			if err != nil {
				return nil, err
			}
			mp = append(mp, [2][]byte{[]byte(k.String()), in})
		}
		slices.SortFunc(mp, func(a, b [2][]byte) int {
			return cmp.Compare(string(a[0]), string(b[0]))
		})
		for _, val := range mp {
			b = append(b, val[1]...)
		}
		return b, nil
	}
	return nil, fmt.Errorf("%w: %T (kind: %v)", ErrUnsupportedType, v, ref.Kind())
}

func marshalBytes(t majorType, sl []byte) ([]byte, error) {
	b := marshalRawInt(t, uint64(len(sl)))
	return append(b, sl...), nil
}

func marshalRawInt(t majorType, i uint64) []byte {
	if i <= 23 {
		return []byte{createType(t, additionalInformation(i))}
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

func marshalKeyVal(key string, val any) ([]byte, error) {
	b, err := marshalBytes(textString, []byte(key))
	if err != nil {
		return nil, err
	}
	k, err := Marshal(val)
	if err != nil {
		return nil, err
	}
	return append(b, k...), nil
}

func structToMap(ref reflect.Value) (map[string]any, error) {
	tp := ref.Type()
	nb := tp.NumField()
	mp := make(map[string]any, nb)
	for i := range nb {
		f := ref.Field(i)
		if !f.CanInterface() {
			continue
		}
		val := f.Interface()
		opts := optionsOf(tp.Field(i))
		if opts.omitempty && reflect.DeepEqual(f.Interface(), reflect.Zero(f.Type()).Interface()) {
			continue
		} else if opts.toString {
			switch cv := val.(type) {
			case fmt.Stringer:
				val = cv.String()
			case encoding.TextMarshaler:
				b, err := cv.MarshalText()
				if err != nil {
					return nil, err
				}
				val = string(b)
			default:
				val = fmt.Sprintf("%v", val)
			}
		}
		mp[opts.name] = val
	}
	return mp, nil
}
