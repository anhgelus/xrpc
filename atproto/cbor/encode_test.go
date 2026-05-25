package cbor

import (
	"slices"
	"testing"
)

func doMarshal(t *testing.T, v any, exp ...byte) {
	res, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res, exp) {
		t.Errorf("bad result for %v: %x, wanted %x", v, res, exp)
	}
}

func TestMarshal_Int(t *testing.T) {
	doMarshal(t, 0, 0)
	doMarshal(t, 1, 1)
	doMarshal(t, 10, 0x0a)
	doMarshal(t, 23, 0x17)
	doMarshal(t, 24, 0x18, 0x18)
	doMarshal(t, 100, 0x18, 0x64)
	doMarshal(t, 1000, 0x19, 0x03, 0xe8)
	doMarshal(t, 1000000, 0x1a, 0x00, 0x0f, 0x42, 0x40)

	doMarshal(t, -1, 0x20)
	doMarshal(t, -2, 0x21)
	doMarshal(t, -10, 0x29)
	doMarshal(t, -100, 0x38, 0x63)
	doMarshal(t, -1000, 0x39, 0x03, 0xe7)
}

func TestMarshal_Bool(t *testing.T) {
	doMarshal(t, false, 0xf4)
	doMarshal(t, true, 0xf5)
	var dummy *int
	doMarshal(t, dummy, 0xf6)
}

func TestMarshal_Bytes(t *testing.T) {
	doMarshal(t, []byte(""), 0x40)
	doMarshal(t, []byte{0x01, 0x02, 0x03}, 0x43, 0x01, 0x02, 0x03)
}

func TestMarshal_String(t *testing.T) {
	doMarshal(t, "", 0x60)
	doMarshal(t, "a", 0x61, 0x61)
	doMarshal(t, "hello", 0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
	doMarshal(t, "IETF", 0x64, 0x49, 0x45, 0x54, 0x46)
	doMarshal(t, "ü", 0x62, 0xc3, 0xbc)
}

func TestMarshal_Array(t *testing.T) {
	doMarshal(t, []bool{}, 0x80)
	doMarshal(t, []int{1, 2, 3}, 0x83, 0x01, 0x02, 0x03)
	doMarshal(t, []bool{true, false}, 0x82, 0xf5, 0xf4)
}

func TestMarshal_Map(t *testing.T) {
	doMarshal(t, map[string]any{}, 0xa0)
	doMarshal(t, map[string]int{"a": 1}, 0xa1, 0x61, 0x61, 0x01)
	doMarshal(t, map[string]int{"a": 1, "b": 2}, 0xa2, 0x61, 0x61, 0x01, 0x61, 0x62, 0x02)
}

func doMarshalStruct(t *testing.T, v any, exp ...any) {
	if len(exp)%2 != 0 {
		t.Fatal("invalid args")
	}
	ln := len(exp) / 2
	cv := make(map[string]any, ln)
	for i := range ln {
		cv[exp[i*2].(string)] = exp[i*2+1]
	}
	mp, err := Marshal(cv)
	if err != nil {
		t.Fatal(err)
	}
	doMarshal(t, v, mp...)
}

func TestMarshal_Struct(t *testing.T) {
	doMarshalStruct(t, struct {
		A uint
		B uint `json:"b"`
		C uint `json:"d" cbor:"c"`
	}{0, 1, 2}, "A", 0, "b", 1, "c", 2)
	doMarshalStruct(t, struct {
		A uint `cbor:",omitempty"`
		B uint `json:"b,omitempty"`
		C uint `cbor:"c,string"`
	}{0, 0, 2}, "c", "2")
	doMarshalStruct(t, struct {
		A uint `cbor:"a,omitempty,string"`
	}{0})
	doMarshalStruct(t, struct {
		A uint `cbor:"a,omitempty,string"`
	}{1}, "a", "1")
}
