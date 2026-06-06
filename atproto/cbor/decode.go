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

var anySpecialCase = reflect.TypeFor[any]()

// Unmarshal a type from CBOR.
// Returns the unparsed data.
func Unmarshal(b []byte, v any) (rest []byte, err error) {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Pointer || ptr.IsNil() {
		return nil, fmt.Errorf("%w: given %T", ErrUnmarshalRequiresPointer, v)
	}
	for ptr.Elem().Kind() == reflect.Pointer {
		if ptr.Elem().IsNil() {
			t := ptr.Elem().Type()
			ref := reflect.New(t.Elem())
			ptr.Elem().Set(ref)
		}
		ptr = ptr.Elem()
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
	r := &ByteReader{bytes: b}
	m, a := extractHead(r)

	switch m {
	case unsignedInt:
		var t uint64
		t, err = unmarshalRawUint(a, r)
		if err != nil {
			return
		}
		err = unmarshalInt(ptr, reflect.ValueOf(t))
	case negativeInt:
		var t uint64
		t, err = unmarshalRawUint(a, r)
		if err != nil {
			return
		}
		err = unmarshalInt(ptr, reflect.ValueOf(int64(t)*-1-1))
	case byteString:
		err = unmarshalBytes(ptr, a, r)
	case textString:
		sub := reflect.New(reflect.TypeFor[string]())
		err = unmarshalBytes(sub, a, r)
		ptr.Elem().Set(sub.Elem().Convert(ptr.Elem().Type()))
	case array:
		err = unmarshalArray(ptr, a, r)
	case mapT:
		err = unmarshalMap(ptr, a, r)
	case tag:
		return nil, fmt.Errorf("%w: must use Tag to decode a tag", ErrInvalidType)
	case simpleValues:
		switch a {
		case 20:
			ptr.Elem().Set(reflect.ValueOf(false))
		case 21:
			ptr.Elem().Set(reflect.ValueOf(true))
		case 22:
			ptr.SetZero()
		default:
			err = fmt.Errorf("%w: unknown simple values, for data [% x]", ErrNotCBOR, b)
		}
	}
	if err != nil {
		return
	}
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

func unmarshalInt(ptr reflect.Value, val reflect.Value) error {
	t := ptr.Elem().Type()
	if t == anySpecialCase {
		ptr.Elem().Set(val)
		return nil
	}
	if val.CanConvert(t) {
		ptr.Elem().Set(val.Convert(t))
		return nil
	}
	return fmt.Errorf("%w: %v is not an (u)int", ErrInvalidType, t.String())
}

func unmarshalBytes(ptr reflect.Value, a additionalInformation, r *ByteReader) error {
	ln, err := unmarshalRawUint(a, r)
	if err != nil {
		return err
	}
	b := reflect.ValueOf(r.More(uint(ln)))
	t := ptr.Elem().Type()
	if t == anySpecialCase {
		ptr.Elem().Set(b)
	} else if b.CanConvert(ptr.Elem().Type()) {
		ptr.Elem().Set(b.Convert(t))
	} else {
		ptr.Elem().Set(b)
	}
	return nil
}

func unmarshalArray(ptr reflect.Value, a additionalInformation, r *ByteReader) error {
	ln, err := unmarshalRawUint(a, r)
	if err != nil {
		return err
	}
	t := ptr.Elem().Type()
	var inner reflect.Value
	switch t.Kind() {
	case reflect.Interface:
		inner = reflect.MakeSlice(reflect.SliceOf(t), int(ln), int(ln))
	case reflect.Array:
		inner = reflect.New(t).Elem()
	default:
		inner = reflect.MakeSlice(t, int(ln), int(ln))
	}
	if ln == 0 {
		ptr.Elem().Set(inner)
		return nil
	}
	for i := range ln {
		var val reflect.Value
		if t == anySpecialCase {
			val = reflect.New(anySpecialCase)
		} else {
			val = reflect.New(t.Elem())
		}
		rest, err := Unmarshal(r.Drain(), val.Interface())
		if err != nil {
			return err
		}
		r.Reset(rest)
		inner.Index(int(i)).Set(val.Elem())
	}
	ptr.Elem().Set(inner)
	return nil
}

func unmarshalKeyVal(r *ByteReader, t reflect.Type) (string, reflect.Value, error) {
	var s string
	rest, err := Unmarshal(r.Drain(), &s)
	if err != nil {
		return "", reflect.Zero(t), err
	}
	ptr := reflect.New(t)
	rest, err = Unmarshal(rest, ptr.Interface())
	if err != nil {
		return "", reflect.Zero(t), err
	}
	r.Reset(rest)
	return s, ptr.Elem(), nil
}

type fieldInfo struct {
	opt   options
	field reflect.Value
	typ   reflect.StructField
}

func unmarshalMap(ptr reflect.Value, a additionalInformation, r *ByteReader) error {
	k := ptr.Elem().Kind()
	isMap := k == reflect.Map
	isAny := ptr.Elem().Type() == anySpecialCase
	if !isMap && k != reflect.Struct && !isAny {
		return fmt.Errorf(
			"%w: must use a Go map or a struct for a CBOR map, not %v",
			ErrInvalidType, k)
	}
	t := ptr.Elem().Type()
	if isMap && t.Key().Kind() != reflect.String {
		return fmt.Errorf("%w: map must use string as keys", ErrInvalidType)
	}
	ln, err := unmarshalRawUint(a, r)
	if err != nil {
		return err
	}
	var mp reflect.Value
	if isMap {
		mp = reflect.MakeMapWithSize(t, int(ln))
	} else {
		mp = reflect.MakeMapWithSize(
			reflect.MapOf(reflect.TypeFor[string](), reflect.TypeFor[any]()),
			int(ln))
	}
	for range ln {
		key, in, err := unmarshalKeyVal(r, mp.Type().Elem())
		if err != nil {
			return err
		}
		mp.SetMapIndex(reflect.ValueOf(key), in)
	}
	if isMap || isAny {
		ptr.Elem().Set(mp)
		return nil
	}
	return unmarshalMapIntoStruct(ptr, mp)
}

func unmarshalMapIntoStruct(ptr reflect.Value, mp reflect.Value) error {
	el := ptr.Elem()
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
		v := val
		if f.opt.toString {
			ptr := reflect.New(f.typ.Type)
			err := unmarshalStringConvertisser(ptr, v)
			if err != nil {
				return err
			}
			v = ptr.Elem()
		}
		t := f.typ.Type
		if v.CanConvert(t) {
			v = v.Convert(t)
		} else if v.Type() != t {
			return fmt.Errorf("%w: cannot convert %v into %v", ErrInvalidType, v.Type(), t)
		}
		if v.IsZero() {
			f.field.SetZero()
		} else {
			f.field.Set(v)
		}
	}
	ptr.Elem().Set(el)
	return nil
}

func unmarshalStringConvertisser(ptr reflect.Value, val reflect.Value) error {
	t := ptr.Elem().Type()
	target := reflect.TypeFor[uint]()
	if t.ConvertibleTo(target) {
		res, err := strconv.ParseUint(val.String(), 10, 64)
		if err == nil {
			ptr.Elem().Set(reflect.ValueOf(res).Convert(target))
			return nil
		}
	}
	target = reflect.TypeFor[int]()
	if t.ConvertibleTo(target) {
		res, err := strconv.ParseInt(val.String(), 10, 64)
		if err == nil {
			ptr.Elem().Set(reflect.ValueOf(res).Convert(target))
			return nil
		}
	}
	target = reflect.TypeFor[bool]()
	if t.ConvertibleTo(target) {
		res, err := strconv.ParseBool(val.String())
		if err == nil {
			ptr.Elem().Set(reflect.ValueOf(res).Convert(target))
			return nil
		}
	}
	return fmt.Errorf(
		"%w: cannot convert %#v (%v) into string",
		ErrInvalidType, val, val.Type(),
	)
}
