package xrpc

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// MapMarshaler is called by [MarshalToMap] when marshaling the type.
type MapMarshaler interface {
	MarshalMap() (any, error)
}

type options struct {
	key string
	fn  func(any) (any, bool)
}

var (
	jsonOpts = []options{
		{"omitempty", func(v any) (any, bool) {
			if v == nil {
				return nil, false
			}
			refVal := reflect.ValueOf(v)
			if reflect.DeepEqual(v, reflect.Zero(refVal.Type()).Interface()) {
				return v, false
			}
			return v, true
		}},
	}
	mapOpts = []options{
		{"string", func(v any) (any, bool) {
			if cv, ok := v.(fmt.Stringer); ok {
				return cv.String(), true
			}
			return fmt.Sprintf("%v", v), true
		}},
	}
)

// MarshalToMap transforms a struct into a map.
// If v is not a struct, it returns its value.
//
// Implements [MapMarshaler] to have a custom behavior.
//
// It supports the json tag:
//
//	type A struct {
//	   Foo string `json:"bar"`
//	}
//
// The field Foo will have the key bar.
// It supports the option omitempty.
// See [encoding/json] for more information.
//
// If you want to convert the value to a string, you add the option
//
//	map:"string"
//
// to the field.
func MarshalToMap(v any) (any, error) {
	// if custom
	if conv, ok := v.(MapMarshaler); ok && v != nil {
		return conv.MarshalMap()
	}
	// if not a struct
	ref := reflect.ValueOf(v)
	switch ref.Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		if ref.IsNil() {
			return nil, nil
		}
		return MarshalToMap(ref.Elem().Interface())
	default:
		return v, nil
	}
	// handling struct
	refType := ref.Type()
	fields := ref.NumField()
	mp := make(map[string]any, fields)
	for i := range fields {
		field := ref.Field(i)
		fieldType := refType.Field(i)
		val, err := MarshalToMap(field.Interface())
		if err != nil {
			return nil, err
		}
		name, val, ok := handleTags(fieldType, fieldType.Name, val)
		if ok {
			mp[name] = val
		}
	}
	// if struct is a record
	if rec, ok := v.(Record); ok {
		mp["$type"] = rec.Type()
	}
	return mp, nil
}

func handleTags(fieldType reflect.StructField, name string, val any) (string, any, bool) {
	data := strings.Split(fieldType.Tag.Get("json"), ",")
	if len(data) == 0 {
		return name, val, true
	}
	if len(data[0]) > 0 {
		name = data[0]
	}
	if name == "-" || (name == "" && len(data) == 1) {
		return name, val, false
	}
	if len(data) > 1 {
		var ok bool
		name, val, ok = applyOpts(data[1:], jsonOpts, name, val)
		if !ok {
			return name, val, ok
		}
	}

	data = strings.Split(fieldType.Tag.Get("map"), ",")
	return applyOpts(data, mapOpts, name, val)
}

func applyOpts(data []string, opts []options, name string, val any) (string, any, bool) {
	ok := true
	for i := 0; i < len(opts) && ok; i++ {
		opt := opts[i]
		if slices.Contains(data, opt.key) {
			val, ok = opt.fn(val)
		}
	}
	return name, val, ok
}
