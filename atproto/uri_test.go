package atproto

import (
	"testing"

	"pgregory.net/rapid"
)

var (
	invalidURI = []string{
		"at://foo.com/",
		"at://user:pass@foo.com",
		"https://example.org",
	}
	invalidURILexicon = []string{
		"at://foo.com/example/123",
		"at://computer",
		"at://example.com:3000",
	}
)

func genURIHandle(t *rapid.T, label string) (URI[Handle], string) {
	domain, _ := genDomain(t, false, label+" handle")
	uri := Handle(domain.Authority).URI()
	if !rapid.Bool().Draw(t, label+" collection?") {
		return uri, uri.String()
	}
	return genURI(t, label, uri)
}

func genURIDID(t *rapid.T, label string) (URI[*DID], string) {
	raw := genDid(t, label+" did")
	did, _ := ParseDID(raw)
	uri := did.URI()
	if !rapid.Bool().Draw(t, label+" collection?") {
		return uri, uri.String()
	}
	return genURI(t, label, uri)
}

func genURI[T Authority](t *rapid.T, label string, uri URI[T]) (URI[T], string) {
	domain, _ := genDomain(t, true, label+" collection")
	uri.collection = domain
	if !rapid.Bool().Draw(t, label+" rkey?") {
		return uri, uri.String()
	}
	raw := genRecordKey().Draw(t, label+" rkey")
	rkey := RecordKey(raw)
	uri.recordKey = &rkey
	return uri, uri.String()
}

func TestParseRawURI(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		u, s := genURIHandle(t, "uri")
		t.Log(s)
		raw, err := ParseRawURI(s)
		if err != nil {
			t.Fatal(err)
		}
		if !raw.IsHandle() {
			t.Fatal("expected handle")
		}
		uri, err := raw.Handle()
		if err != nil {
			t.Fatal(err)
		}
		if uri.authority != u.authority {
			t.Errorf("invalid authority: %s, wanted %s", uri.authority, u.authority)
		}
		if u.collection != nil {
			if *uri.collection != *u.collection {
				t.Errorf("invalid collection: %s, wanted %s", uri.collection, u.collection)
			}
		}
		if u.recordKey != nil {
			if *uri.recordKey != *u.recordKey {
				t.Errorf("invalid rkey: %s, wanted %s", *uri.recordKey, *u.recordKey)
			}
		}
	})
	rapid.Check(t, func(t *rapid.T) {
		u, s := genURIDID(t, "uri")
		t.Log(s)
		raw, err := ParseRawURI(s)
		if err != nil {
			t.Fatal(err)
		}
		if !raw.IsDID() {
			t.Fatal("expected DID")
		}
		uri, err := raw.DID()
		if err != nil {
			t.Fatal(err)
		}
		if *uri.authority != *u.authority {
			t.Errorf("invalid authority: %s, wanted %s", uri.authority, u.authority)
		}
		if u.collection != nil {
			if *uri.collection != *u.collection {
				t.Errorf("invalid collection: %s, wanted %s", uri.collection, u.collection)
			}
		}
		if u.recordKey != nil {
			if *uri.recordKey != *u.recordKey {
				t.Errorf("invalid rkey: %s, wanted %s", *uri.recordKey, *u.recordKey)
			}
		}
	})
	fn := func(u string) {
		uri, err := ParseRawURI(u)
		if err == nil {
			if uri.IsDID() {
				_, err = uri.DID()
				if err == nil {
					t.Errorf("expected error for %s", u)
				}
			} else if uri.IsHandle() {
				_, err = uri.Handle()
				if err == nil {
					t.Errorf("expected error for %s", u)
				}
			}
		}
	}
	for _, u := range invalidURI {
		fn(u)
	}
	for _, u := range invalidURILexicon {
		fn(u)
	}
}

func TestURI_Set(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := genDid(t, "did")
		did, _ := ParseDID(raw)
		u1 := did.URI()
		nsid, _ := genDomain(t, true, "collection")
		u2 := u1.SetCollection(nsid)
		if u1.String() == u2.String() {
			t.Errorf("u1 is u2: %s, %s", u1.String(), u2.String())
		}
		if *u2.Authority() != *did {
			t.Errorf("invalid authority: %s, wanted %s", u2.Authority(), did)
		}
		if *u2.Collection() != *nsid {
			t.Errorf("invalid nsid: %s, wanted %s", u2.Collection(), nsid)
		}
		rkey := genRecordKey().Draw(t, "rkey")
		u3 := u2.SetRecordKey(RecordKey(rkey))
		if u2.String() == u3.String() {
			t.Errorf("u2 is u3: %s, %s", u2.String(), u3.String())
		}
		if *u3.Authority() != *did {
			t.Errorf("invalid authority: %s, wanted %s", u3.Authority(), did)
		}
		if *u3.Collection() != *nsid {
			t.Errorf("invalid nsid: %s, wanted %s", u3.Collection(), nsid)
		}
		if *u3.RecordKey() != RecordKey(rkey) {
			t.Errorf("invalid rkey: %s, wanted %s", u3.RecordKey(), rkey)
		}
	})
}
