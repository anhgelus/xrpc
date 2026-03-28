package xrpc

import (
	"fmt"
	"reflect"
	"testing"
)

func validMap(t *testing.T, check any, res map[string]any) {
	mp, ok := check.(map[string]any)
	if !ok {
		t.Errorf("cannot convert %v into a map", check)
		return
	}
	if len(mp) != len(res) {
		t.Errorf("invalid value: got %v, wanted %v", check, res)
	}
	for k, val := range mp {
		if !reflect.DeepEqual(res[k], val) {
			t.Errorf("invalid value at %s, got %v, wanted %v", k, val, res[k])
		}
	}
}

type test1 struct {
	A string `json:"a"`
	B string `json:",omitempty"`
}

func TestMarshalToMap_Simple(t *testing.T) {
	h := test1{"aaa", ""}
	v, err := MarshalToMap(h)
	if err != nil {
		t.Fatal(err)
	}
	validMap(t, v, map[string]any{"a": "aaa"})

	h = test1{"", "bbb"}
	v, err = MarshalToMap(h)
	if err != nil {
		t.Fatal(err)
	}
	validMap(t, v, map[string]any{"a": "", "B": "bbb"})
}

type test2 struct {
	Hey *test1 `json:"hey,omitempty"`
	Key string `json:"key"`
}

func TestMarshalToMap_Nested(t *testing.T) {
	h := test2{&test1{"aaa", ""}, "key"}
	v, err := MarshalToMap(h)
	if err != nil {
		t.Fatal(err)
	}
	validMap(t, v, map[string]any{"hey": map[string]any{"a": "aaa"}, "key": "key"})

	h = test2{nil, "k"}
	v, err = MarshalToMap(h)
	if err != nil {
		t.Fatal(err)
	}
	validMap(t, v, map[string]any{"key": "k"})
}

func (t test1) String() string {
	return fmt.Sprintf("%s:%s", t.A, t.B)
}

type test3 struct {
	Default int `json:"default"`
	ConvDef int `json:"conv" map:"string"`
}

func TestMarshalToMap_String(t *testing.T) {
	h := test3{1, 1}
	v, err := MarshalToMap(h)
	if err != nil {
		t.Fatal(err)
	}
	validMap(t, v, map[string]any{"default": 1, "conv": "1"})
}

type test4 struct {
	A int
}

func (t test4) MarshalMap() (any, error) {
	return map[string]any{"a": t.A + 5}, nil
}

func TestMarshalToMap_CustomMarshal(t *testing.T) {
	h := test4{0}
	v, err := MarshalToMap(h)
	if err != nil {
		t.Fatal(err)
	}
	validMap(t, v, map[string]any{"a": 5})
}
