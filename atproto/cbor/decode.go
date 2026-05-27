package cbor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	ErrUnmarshalRequiresPointer = errors.New("unmarshal requires a pointer")
	ErrUnsupportedHead          = errors.New("unsupported head")
	ErrInvalidType              = errors.New("invalid type")
	ErrNotCBOR                  = errors.New("not a CBOR")
)

// Unmarshaler describes how the type is decoded from CBOR.
type Unmarshaler interface {
	// UnmarshalCBOR returns the unparsed bytes.
	UnmarshalCBOR(b []byte) ([]byte, error)
}

func extractHead(r *ByteReader) (majorType, additionalInformation) {
	b := r.Next()
	m := b & 0b111_00000
	a := b & 0b000_11111
	return majorType(m >> 5), additionalInformation(a)
}

// Unmarshal a type from CBOR.
// Returns the unparsed data.
func Unmarshal(b []byte, v any) (rest []byte, err error) {
	ref := reflect.ValueOf(v)
	if ref.Kind() != reflect.Pointer || ref.IsNil() {
		return nil, fmt.Errorf("%w: given %T", ErrUnmarshalRequiresPointer, v)
	}
	for ref.Elem().Kind() == reflect.Pointer {
		ref = ref.Elem()
	}
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%w: %v, for data [% x]", ErrNotCBOR, v, b)
		}
	}()
	switch cv := v.(type) {
	case Unmarshaler:
		return cv.UnmarshalCBOR(b)
	case Tag:
		return unmarshalTag(b, cv)
	}
	r := &ByteReader{Bytes: b}
	m, a := extractHead(r)

	var val any
	switch m {
	case unsignedInt:
		var t uint64
		t, err = unmarshalRawUint(a, r)
		if err != nil {
			return nil, err
		}
		val, err = unmarshalInt(ref.Elem().Type(), reflect.ValueOf(t))
	case negativeInt:
		var t uint64
		t, err = unmarshalRawUint(a, r)
		if err != nil {
			return nil, err
		}
		val, err = unmarshalInt(ref.Elem().Type(), reflect.ValueOf(int64(t)*-1-1))
	case byteString:
		val, err = unmarshalBytes(a, r)
	case textString:
		var t []byte
		t, err = unmarshalBytes(a, r)
		if err == nil {
			val = string(t)
		}
	case array:
		val, err = unmarshalArray(ref.Elem(), a, r)
	case mapT:
		k := ref.Elem().Kind()
		isMap := k == reflect.Map
		if !isMap && k != reflect.Struct {
			return nil, fmt.Errorf(
				"%w: must use a Go map or a struct for a CBOR map, not %v",
				ErrInvalidType, k,
			)
		}
		t := ref.Elem().Type()
		if isMap && t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("%w: map must use string as keys", ErrInvalidType)
		}
		var ln uint64
		ln, err = unmarshalRawUint(a, r)
		if err != nil {
			return
		}
		var mp reflect.Value
		if isMap {
			mp = reflect.MakeMapWithSize(t, int(ln))
		} else {
			mp = reflect.MakeMapWithSize(
				reflect.MapOf(reflect.TypeFor[string](), reflect.TypeFor[any]()),
				int(ln),
			)
		}
		for range ln {
			var key string
			var in reflect.Value
			key, in, err = unmarshalKeyVal(r, mp.Type().Elem())
			if err != nil {
				return
			}
			mp.SetMapIndex(reflect.ValueOf(key), in)
		}
		if k == reflect.Map {
			val = mp.Interface()
		} else {
			val, err = unmarshalMapIntoStruct(mp, ref)
		}
	case tag:
		return nil, fmt.Errorf("%w: must use a Tag to decode a tag", ErrInvalidType)
	case simpleValues:
		switch a {
		case 20:
			val = false
		case 21:
			val = true
		case 22:
			return r.Drain(), nil
		default:
			err = fmt.Errorf("%w: unknown simple values, for data [% x]", ErrNotCBOR, b)
		}
	}
	if err != nil {
		return nil, err
	}
	ref.Elem().Set(reflect.ValueOf(val))
	return r.Drain(), nil
}

func unmarshalRawUint(a additionalInformation, r *ByteReader) (uint64, error) {
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

func unmarshalInt(t reflect.Type, val reflect.Value) (any, error) {
	if val.CanConvert(t) {
		return val.Convert(t).Interface(), nil
	}
	if t.Kind() == reflect.Interface {
		return val.Interface(), nil
	}
	return nil, fmt.Errorf("%w: %v is not an (u)int", ErrInvalidType, t.String())
}

func unmarshalBytes(a additionalInformation, r *ByteReader) ([]byte, error) {
	ln, err := unmarshalRawUint(a, r)
	if err != nil {
		return nil, err
	}
	return r.More(uint(ln)), nil
}

func unmarshalArray(val reflect.Value, a additionalInformation, r *ByteReader) (any, error) {
	ln, err := unmarshalRawUint(a, r)
	if err != nil {
		return nil, err
	}
	t := val.Type().Elem()
	for range ln {
		ptr := reflect.New(t)
		rest, err := Unmarshal(r.Bytes[r.I:], ptr.Interface())
		if err != nil {
			return nil, err
		}
		r.Bytes = rest
		r.I = 0
		val = reflect.Append(val, ptr.Elem())
	}
	return val.Interface(), nil
}

func unmarshalKeyVal(r *ByteReader, t reflect.Type) (string, reflect.Value, error) {
	var s string
	rest, err := Unmarshal(r.Bytes[r.I:], &s)
	if err != nil {
		return "", reflect.ValueOf(nil), err
	}
	ptr := reflect.New(t)
	rest, err = Unmarshal(rest, ptr.Interface())
	if err != nil {
		return "", reflect.ValueOf(nil), err
	}
	r.Bytes = rest
	r.I = 0
	return s, ptr.Elem(), err
}

type fieldInfo struct {
	opt   options
	field reflect.Value
	typ   reflect.StructField
}

func unmarshalMapIntoStruct(mp reflect.Value, v reflect.Value) (any, error) {
	el := v.Elem()
	fields := make(map[string]fieldInfo, el.NumField())
	for i := range el.NumField() {
		f := el.Type().Field(i)
		if !f.IsExported() {
			continue
		}
		opt := optionsOf(f)
		fields[opt.name] = fieldInfo{opt, el.Field(i), f}
	}
	for k, val := range mp.Seq2() {
		f, ok := fields[k.String()]
		if !ok {
			continue
		}
		if val.Kind() == reflect.Interface {
			val = reflect.ValueOf(val.Interface())
		}
		if f.opt.omitempty && reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type())) {
			continue
		}
		if f.opt.toString && val.Kind() == reflect.String {
			var err error
			val, err = unmarshalStringConvertisser(f.typ.Type, val)
			if err != nil {
				return nil, err
			}
		}
		t := f.field.Type()
		if val.CanConvert(t) {
			val = val.Convert(t)
		} else if val.Type() != t {
			return nil, fmt.Errorf("%w: cannot convert %v into %v", ErrInvalidType, val.Type(), t)
		}
		if val.IsZero() {
			f.field.SetZero()
		} else {
			f.field.Set(val)
		}
	}
	return el.Interface(), nil
}

func unmarshalStringConvertisser(typ reflect.Type, val reflect.Value) (reflect.Value, error) {
	if typ.ConvertibleTo(reflect.TypeFor[uint]()) {
		res, err := strconv.ParseUint(val.String(), 10, 64)
		if err == nil {
			return reflect.ValueOf(res), nil
		}
	}
	if typ.ConvertibleTo(reflect.TypeFor[int]()) {
		res, err := strconv.ParseInt(val.String(), 10, 64)
		if err == nil {
			return reflect.ValueOf(res), nil
		}
	}
	if typ.ConvertibleTo(reflect.TypeFor[bool]()) {
		res, err := strconv.ParseBool(val.String())
		if err == nil {
			val = reflect.ValueOf(res)
			return reflect.ValueOf(res), nil
		}
	}
	return reflect.Zero(typ), fmt.Errorf(
		"%w: cannot convert %#v (%T) into string",
		ErrInvalidType, val.Interface(), val.Interface(),
	)
}
