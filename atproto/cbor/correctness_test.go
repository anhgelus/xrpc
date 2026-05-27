package cbor

import (
	"fmt"
	"maps"
	"testing"

	"pgregory.net/rapid"
)

func testMapCorrectness[T comparable](t *testing.F, gen *rapid.Generator[T]) {
	t.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		mp := rapid.MapOf(rapid.String(), gen).Draw(t, fmt.Sprintf("map: %T", map[string]T{}))
		b, err := Marshal(mp)
		if err != nil {
			t.Fatal(err)
		}
		var res map[string]T
		rest, err := Unmarshal(b, &res)
		if err != nil {
			t.Fatal(err)
		}
		if len(rest) != 0 {
			t.Errorf("expected no rest: % x", rest)
		}
		if !maps.Equal(mp, res) {
			t.Errorf("expected equals maps:\ninput %#v\noutput %#v", mp, res)
		}
	}))
}

func FuzzCorrectness_String(t *testing.F) {
	testMapCorrectness(t, rapid.String())
}

func FuzzCorrectness_Int(t *testing.F) {
	testMapCorrectness(t, rapid.Int())
}

func FuzzCorrectness_Uint(t *testing.F) {
	testMapCorrectness(t, rapid.Uint())
}
