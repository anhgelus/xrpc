package cbor

import (
	"reflect"
	"testing"
)

func doUnmarshal[T any](t *testing.T, v []byte, exp T) {
	var raw T
	rest, err := Unmarshal(v, &raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 0 {
		t.Errorf("not everything was parsed: % x", rest)
	}
	if !reflect.DeepEqual(raw, exp) {
		t.Errorf("bad result for % x: %#v, wanted %#v", v, raw, exp)
	}
}

func TestUnmarshal_Int(t *testing.T) {
	doUnmarshal(t, []byte{0}, 0)
	doUnmarshal(t, []byte{0x01}, 1)
	doUnmarshal(t, []byte{0x0a}, 10)
	doUnmarshal(t, []byte{0x17}, 23)
	doUnmarshal(t, []byte{0x18, 0x18}, 24)
	doUnmarshal(t, []byte{0x18, 0x64}, 100)
	doUnmarshal(t, []byte{0x19, 0x03, 0xe8}, 1000)
	doUnmarshal(t, []byte{0x1a, 0x00, 0x0f, 0x42, 0x40}, 1000000)

	doUnmarshal(t, []byte{0x20}, -1)
	doUnmarshal(t, []byte{0x21}, -2)
	doUnmarshal(t, []byte{0x29}, -10)
	doUnmarshal(t, []byte{0x38, 0x63}, -100)
	doUnmarshal(t, []byte{0x39, 0x03, 0xe7}, -1000)
}

func TestUnmarshal_Bool(t *testing.T) {
	doUnmarshal(t, []byte{0xf4}, false)
	doUnmarshal(t, []byte{0xf5}, true)
	var dummy *bool
	doUnmarshal(t, []byte{0xf6}, dummy)
}

func TestUnmarshal_Bytes(t *testing.T) {
	doUnmarshal(t, []byte{0x40}, []byte{})
	doUnmarshal(t, []byte{0x43, 0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03})
}

func TestUnmarshal_String(t *testing.T) {
	doUnmarshal(t, []byte{0x60}, "")
	doUnmarshal(t, []byte{0x61, 0x61}, "a")
	doUnmarshal(t, []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, "hello")
	doUnmarshal(t, []byte{0x64, 0x49, 0x45, 0x54, 0x46}, "IETF")
	doUnmarshal(t, []byte{0x62, 0xc3, 0xbc}, "ü")
}

func TestUnmarshal_Array(t *testing.T) {
	doUnmarshal(t, []byte{0x80}, []byte(nil))
	doUnmarshal(t, []byte{0x83, 0x01, 0x02, 0x03}, []int{1, 2, 3})
	doUnmarshal(t, []byte{0x82, 0xf5, 0xf4}, []bool{true, false})
}

func TestUnmarshal_Map(t *testing.T) {
	doUnmarshal(t, []byte{0xa0}, map[string]any{})
	doUnmarshal(t, []byte{0xa1, 0x61, 0x61, 0x01}, map[string]int{"a": 1})
	doUnmarshal(t, []byte{0xa2, 0x61, 0x61, 0x01, 0x61, 0x62, 0x02}, map[string]int{"a": 1, "b": 2})
}

func doUnmarshalStruct[T any](t *testing.T, exp T, v ...any) {
	if len(v)%2 != 0 {
		t.Fatal("invalid args")
	}
	ln := len(v) / 2
	cv := make(map[string]any, ln)
	for i := range ln {
		cv[v[i*2].(string)] = v[i*2+1]
	}
	mp, err := Marshal(cv)
	if err != nil {
		t.Fatal(err)
	}
	doUnmarshal(t, mp, exp)
}

func TestUnmarshal_Struct(t *testing.T) {
	doUnmarshalStruct(t, struct {
		A uint
		B uint `json:"b"`
		C uint `json:"d" cbor:"c"`
	}{0, 1, 2}, "A", 0, "b", 1, "c", 2)
	doUnmarshalStruct(t, struct {
		A uint `cbor:",omitempty"`
		B uint `json:"b,omitempty"`
		C uint `cbor:"c,string"`
	}{0, 0, 2}, "c", "2")
	doUnmarshalStruct(t, struct {
		A uint `cbor:"a,omitempty,string"`
	}{0})
	doUnmarshalStruct(t, struct {
		A uint `cbor:"a,omitempty,string"`
	}{1}, "a", "1")
}
