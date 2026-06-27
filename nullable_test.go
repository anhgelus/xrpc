package xrpc_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"anhgelus.world/xrpc"
	"pgregory.net/rapid"
)

func testNullable[T any](t *rapid.T, gen *rapid.Generator[T]) {
	var n xrpc.Nullable[T]
	val := gen.Draw(t, "content")
	n.Set(&val)
	if !n.Present() {
		t.Fatal("value not present")
	}
	b, err := xrpc.Marshal(&n, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v -> %s\n", n, b)
	var res *T
	err = json.Unmarshal(b, &res)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(n.Get(), res) {
		t.Errorf("invalid nullable: wanted %v, not %v", n.Get(), res)
	}
}

func testNull[T any](t *rapid.T) {
	var n xrpc.Nullable[T]
	if n.Present() {
		t.Fatal("value present")
	}
	b, err := xrpc.Marshal(&n, false)
	if err != nil {
		t.Fatal(err)
	}
	var res *T
	err = json.Unmarshal(b, &res)
	if err != nil {
		t.Fatal(err)
	}
	if res != nil {
		t.Errorf("invalid nullable: wanted nil, not %v", res)
	}
}

func TestNullable_Encoding(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		testNullable(t, rapid.Int())
		testNull[int](t)
	})
	rapid.Check(t, func(t *rapid.T) {
		testNullable(t, rapid.Uint())
		testNull[uint](t)
	})
	rapid.Check(t, func(t *rapid.T) {
		testNullable(t, rapid.String())
		testNull[string](t)
	})
	rapid.Check(t, func(t *rapid.T) {
		testNullable(t, rapid.Bool())
		testNull[bool](t)
	})
}
