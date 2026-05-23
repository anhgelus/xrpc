package cbor

import (
	"slices"
	"testing"
)

func doMarshal(t *testing.T, v any, exp []byte) {
	res, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res, exp) {
		t.Errorf("bad result for %v: %x, wanted %x", v, res, exp)
	}
}

func TestMarshal_Int(t *testing.T) {
	doMarshal(t, 0, []byte{0})
	doMarshal(t, 1, []byte{1})
	doMarshal(t, 10, []byte{0x0a})
	doMarshal(t, 23, []byte{0x17})
	doMarshal(t, 24, []byte{0x18, 0x18})
	doMarshal(t, 100, []byte{0x18, 0x64})
	doMarshal(t, 1000, []byte{0x19, 0x03, 0xe8})
	doMarshal(t, 1000000, []byte{0x1a, 0x00, 0x0f, 0x42, 0x40})

	doMarshal(t, -1, []byte{0x20})
	doMarshal(t, -2, []byte{0x21})
	doMarshal(t, -10, []byte{0x29})
	doMarshal(t, -100, []byte{0x38, 0x63})
	doMarshal(t, -1000, []byte{0x39, 0x03, 0xe7})
}

func TestMarshal_Bool(t *testing.T) {
	doMarshal(t, false, []byte{0xf4})
	doMarshal(t, true, []byte{0xf5})
	var dummy *int
	doMarshal(t, dummy, []byte{0xf6})
}

func TestMarshal_Bytes(t *testing.T) {
	doMarshal(t, []byte(""), []byte{0x40})
	doMarshal(t, []byte{0x01, 0x02, 0x03}, []byte{0x43, 0x01, 0x02, 0x03})
}

func TestMarshal_String(t *testing.T) {
	doMarshal(t, "", []byte{0x60})
	doMarshal(t, "a", []byte{0x61, 0x61})
	doMarshal(t, "hello", []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f})
	doMarshal(t, "IETF", []byte{0x64, 0x49, 0x45, 0x54, 0x46})
	doMarshal(t, "ü", []byte{0x62, 0xc3, 0xbc})
}
