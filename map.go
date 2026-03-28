package xrpc

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// MarshalerMap implements a custom [MarshalToMap].
type MarshalerMap interface {
	MarshalMap() (any, error)
}

type options struct {
	key string
	fn  func(any) (any, bool)
}

var jsonOpts = []options{
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

var mapOpts = []options{
	{"string", func(v any) (any, bool) {
		if cv, ok := v.(fmt.Stringer); ok {
			return cv.String(), true
		}
		return fmt.Sprintf("%v", v), true
	}},
}

func getElem(v reflect.Value) (any, error) {
	if !v.CanInterface() {
		return nil, nil
	}
	val := v.Interface()
	if conv, ok := val.(MarshalerMap); ok {
		return conv.MarshalMap()
	}
	switch v.Kind() {
	case reflect.Struct:
		return MarshalToMap(val)
	case reflect.Pointer:
		if v.IsNil() {
			return nil, nil
		}
		return getElem(v.Elem())
	default:
		return val, nil
	}
}

// MarshalToMap transforms a struct into a map.
//
// If v is not a struct, it returns its value.
//
// Implements [MarshalerMap] to have a custom behavior.
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
	if conv, ok := v.(MarshalerMap); ok {
		return conv.MarshalMap()
	}
	refType := ref.Type()
	fields := ref.NumField()
	mp := make(map[string]any, fields)
	for i := range fields {
		field := ref.Field(i)
		fieldType := refType.Field(i)
		val, err := getElem(field)
		if err != nil {
			return nil, err
		}
		name, val, ok := handleTags(fieldType, fieldType.Name, val)
		if ok {
			mp[name] = val
		}
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
		tagOpts := data[1:]
		ok := true
		for i := 0; i < len(jsonOpts) && ok; i++ {
			opt := jsonOpts[i]
			if slices.Contains(tagOpts, opt.key) {
				val, ok = opt.fn(val)
			}
		}
		if !ok {
			return name, val, false
		}
	}

	data = strings.Split(fieldType.Tag.Get("map"), ",")
	ok := true
	for i := 0; i < len(mapOpts) && ok; i++ {
		opt := mapOpts[i]
		if slices.Contains(data, opt.key) {
			val, ok = opt.fn(val)
		}
	}
	return name, val, ok
}
