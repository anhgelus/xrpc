package cbor

import (
	"reflect"
	"testing"

	"pgregory.net/rapid"
)

func genMap() *rapid.Generator[map[string]any] {
	return rapid.MapOf(rapid.String(), genRandom())
}

func genRandom() *rapid.Generator[any] {
	return rapid.Custom(func(t *rapid.T) any {
		switch rapid.UintMax(6).Draw(t, "val kind") {
		case 0:
			return rapid.Bool().Draw(t, "val")
		case 1:
			return rapid.Uint().Draw(t, "val")
		case 2:
			return rapid.Int().Draw(t, "val")
		case 3:
			return rapid.SliceOf(genRandom()).Draw(t, "val")
		case 4:
			return rapid.Byte().Draw(t, "val")
		case 5:
			return rapid.String().Draw(t, "val")
		case 6:
			return genMap().Draw(t, "val")
		default:
			panic("impossible")
		}
	})
}

func eqMap(mp1 any, mp2 any) bool {
	v1 := reflect.ValueOf(mp1)
	v2 := reflect.ValueOf(mp2)
	if v1.Kind() != v2.Kind() {
		tt := v2.Type()
		if !v1.Type().ConvertibleTo(tt) {
			return false
		}
		v1 = v1.Convert(tt)
	}
	switch v1.Kind() {
	case reflect.Map:
		r := v1.MapRange()
		for r.Next() {
			val, ok := mp2.(map[string]any)[r.Key().String()]
			if !ok {
				return false
			}
			if !eqMap(val, r.Value().Interface()) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if v1.Len() != v2.Len() {
			return false
		}
		for i := range v1.Len() {
			if !eqMap(v1.Index(i).Interface(), v2.Index(i).Interface()) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(v1.Interface(), v2.Interface())
}

func FuzzCorrectness(t *testing.F) {
	t.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		mp := genMap().Draw(t, "map out")
		b, err := Marshal(mp)
		if err != nil {
			t.Fatal(err)
		}
		var res map[string]any
		rest, err := Unmarshal(b, &res)
		if err != nil {
			t.Fatal(err)
		}
		if len(rest) != 0 {
			t.Errorf("expected no rest: % x", rest)
		}
		if !eqMap(mp, res) {
			t.Errorf("expected equals maps:\ninput %#v\noutput %#v", mp, res)
		}
	}))
}
