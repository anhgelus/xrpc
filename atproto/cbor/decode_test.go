package cbor

import (
	"slices"
	"testing"
)

func doUnmarshal[T any](t *testing.T, v []byte, exp T, eq func(a, b T) bool) {
	var raw T
	rest, err := Unmarshal(v, &raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 0 {
		t.Errorf("not everything was parsed: %v", rest)
	}
	if !eq(raw, exp) {
		t.Errorf("bad result for %x: %v, wanted %v", v, raw, exp)
	}
}

func TestUnmarshal_Int(t *testing.T) {
	eq1 := func(a, b uint) bool { return a == b }
	doUnmarshal(t, []byte{0}, 0, eq1)
	doUnmarshal(t, []byte{0x01}, 1, eq1)
	doUnmarshal(t, []byte{0x0a}, 10, eq1)
	doUnmarshal(t, []byte{0x17}, 23, eq1)
	doUnmarshal(t, []byte{0x18, 0x18}, 24, eq1)
	doUnmarshal(t, []byte{0x18, 0x64}, 100, eq1)
	doUnmarshal(t, []byte{0x19, 0x03, 0xe8}, 1000, eq1)
	doUnmarshal(t, []byte{0x1a, 0x00, 0x0f, 0x42, 0x40}, 1000000, eq1)

	eq2 := func(a, b int) bool { return a == b }
	doUnmarshal(t, []byte{0x20}, -1, eq2)
	doUnmarshal(t, []byte{0x21}, -2, eq2)
	doUnmarshal(t, []byte{0x29}, -10, eq2)
	doUnmarshal(t, []byte{0x38, 0x63}, -100, eq2)
	doUnmarshal(t, []byte{0x39, 0x03, 0xe7}, -1000, eq2)
}

func TestUnmarshal_Bool(t *testing.T) {
	doUnmarshal(t, []byte{0xf4}, false, func(a, b bool) bool { return a == b })
	doUnmarshal(t, []byte{0xf5}, true, func(a, b bool) bool { return a == b })
	var dummy *int
	doUnmarshal(t, []byte{0xf6}, dummy, func(a, b *int) bool { return a == b })
}

func TestUnmarshal_Bytes(t *testing.T) {
	doUnmarshal(t, []byte{0x40}, []byte{}, slices.Equal)
	doUnmarshal(t, []byte{0x43, 0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}, slices.Equal)
}

func TestUnmarshal_String(t *testing.T) {
	eq := func(a, b string) bool { return a == b }
	doUnmarshal(t, []byte{0x60}, "", eq)
	doUnmarshal(t, []byte{0x61, 0x61}, "a", eq)
	doUnmarshal(t, []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, "hello", eq)
	doUnmarshal(t, []byte{0x64, 0x49, 0x45, 0x54, 0x46}, "IETF", eq)
	doUnmarshal(t, []byte{0x62, 0xc3, 0xbc}, "ü", eq)
}

func TestUnmarshal_Array(t *testing.T) {
	doUnmarshal(t, []byte{0x80}, []byte{}, slices.Equal)
	doUnmarshal(t, []byte{0x83, 0x01, 0x02, 0x03}, []int{1, 2, 3}, slices.Equal)
	doUnmarshal(t, []byte{0x82, 0xf5, 0xf4}, []bool{true, false}, slices.Equal)
}
