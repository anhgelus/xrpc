package atproto

import (
	"context"
	"errors"
	"net/http"
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
	realDids = []string{
		"did:plc:vtqucb4iga7b5wzza3zbz4so",
		"did:plc:oisofpd7lj26yvgiivf3lxsi",
		"did:plc:w7x22x56pgtta23uulbcahbo",
	}
)

func TestAsDidMethod(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		method := genDidMethod().Draw(t, "method")
		t.Log(method)
		_, v := asDIDMethod(method)
		if !v {
			t.Error("invalid method verification")
		}
	})
}

func genDid(t *rapid.T, label string) string {
	method := rapid.SampledFrom([]DIDMethod{DIDWeb, DIDPlc}).Draw(t, label+" method")
	idsFirst := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCEDFGHIJKLMNOPQRSTUVWXYZ1234567890._:%-"))
	idsLast := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCEDFGHIJKLMNOPQRSTUVWXYZ1234567890._-"))
	identifier := rapid.StringOfN(idsFirst, -1, -1, DIDIdentifierMaxLength-5-len(method)).Draw(t, label+" id first") +
		rapid.StringOfN(idsLast, 1, -1, 1).Draw(t, "id last")
	return "did:" + method.String() + ":" + identifier
}

func TestParseDid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		did := genDid(t, "did")
		t.Log(did)
		d, err := ParseDID(did)
		if err != nil {
			t.Fatal(err)
		}
		if d.String() != did {
			t.Errorf("invalid parsing: %s, wanted %s", d, did)
		}
	})
	for _, did := range invalidDids {
		_, err := ParseDID(did)
		if err == nil {
			t.Errorf("expected error for %s", did)
		}
	}
	for _, did := range invalidDidMethods {
		_, err := ParseDID(did)
		if err == nil {
			t.Errorf("expected error for %s", did)
		} else if ok := errors.Is(err, ErrUnsupportedDIDMethod); !ok {
			t.Errorf("invalid error for %s: %v, wanted %v", did, err, ErrUnsupportedDIDMethod)
		}
	}
}

func TestDid_Document(t *testing.T) {
	if testing.Short() {
		t.Skip("not doing http requests in short")
	}
	for _, d := range realDids {
		t.Log(d)
		did, err := ParseDID(d)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := did.document(context.Background(), http.DefaultClient)
		if err != nil {
			t.Fatal(err)
		}
		for _, as := range doc.AlsoKnownAs {
			t.Log(as)
		}
		if *doc.DID != *did {
			t.Errorf("invalid did resolved: %s, wanted %s", doc.DID, did)
		}
	}
}
