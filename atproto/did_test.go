package atproto

import (
	"errors"
	"testing"

	"pgregory.net/rapid"
)

var rapidLowerCase = rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyz"))

func genDidMethod() *rapid.Generator[string] {
	return rapid.StringOfN(rapidLowerCase, 1, -1, -1)
}

var (
	invalidDids = []string{
		"did:METHOD:val",
		"did:m123:val",
		"DID:method:val",
		"did:method:",
		"did:method:val/two",
		"did:method:val?two",
		"did:method:val#two",
	}
	invalidDidMethods = []string{
		"did:method:val:two",
		"did:m:v",
		"did:method::::val",
		"did:method:-:_:.",
		"did:key:zQ3shZc2QzApp2oymGvQbzP8eKheVshBHbU4ZYjeXqwSKEn6N",
	}
)

func TestAsDidMethod(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		method := genDidMethod().Draw(t, "method")
		_, v := asDidMethod(method)
		if !v {
			t.Error("invalid method verification")
		}
	})
}

func genDid(t *rapid.T, label string) string {
	method := rapid.SampledFrom([]DidMethod{DidWeb, DidPLC}).Draw(t, " method")
	idsFirst := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCEDFGHIJKLMNOPQRSTUVWXYZ1234567890._:%-"))
	idsLast := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCEDFGHIJKLMNOPQRSTUVWXYZ1234567890._%-"))
	identifier := rapid.StringOfN(idsFirst, -1, -1, DidIdentifierMaxLength-5-len(method)).Draw(t, label+" id first") +
		rapid.StringOfN(idsLast, 1, -1, 1).Draw(t, "id last")
	return "did:" + method.String() + ":" + identifier
}

func TestParseDid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		did := genDid(t, "did")
		d, err := ParseDID(did)
		if err != nil {
			t.Fatal(err)
		}
		if d.String() != did {
			t.Errorf("invalid parsing: %s, wanted %s", d, did)
		}
	})
	for _, did := range invalidDids {
		t.Log(did)
		_, err := ParseDID(did)
		if err == nil {
			t.Fatalf("expected error for %s", did)
		}
		t.Log(err)
	}
	for _, did := range invalidDidMethods {
		t.Log(did)
		_, err := ParseDID(did)
		if err == nil {
			t.Fatalf("expected error for %s", did)
		}
		if ok := errors.Is(err, ErrUnsupportedDidMethod); !ok {
			t.Errorf("invalid error: %v, wanted %v", err, ErrUnsupportedDidMethod)
		}
	}
}
