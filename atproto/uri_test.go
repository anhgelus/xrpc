package atproto

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

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
	validURI = []string{
		"at://anhgelus.world",
		"at://hailey.at",
		"at://mewsse.pet",
	}
)

func genURIHandle(t *rapid.T, label string) (RawURI, string) {
	domain, _ := genDomain(t, false, label+" handle")
	var sb strings.Builder
	sb.WriteString("at://")
	sb.WriteString(domain.Authority)
	var r RawURI
	if !rapid.Bool().Draw(t, label+" collection?") {
		return r, sb.String()
	}
	sb.WriteRune('/')
	domain, _ = genDomain(t, true, label+" collection")
	sb.WriteString(domain.String())
	if !rapid.Bool().Draw(t, label+" rkey?") {
		return r, sb.String()
	}
	sb.WriteRune('/')
	raw := genRecordKey().Draw(t, label+" rkey")
	sb.WriteString(raw)
	return r, sb.String()
}

func genURI(t *rapid.T, label string) (URI, string) {
	raw := genDid(t, label+" did")
	did, _ := ParseDID(raw)
	uri := did.URI()
	if !rapid.Bool().Draw(t, label+" collection?") {
		return uri, uri.String()
	}
	domain, _ := genDomain(t, true, label+" collection")
	uri.collection = domain
	if !rapid.Bool().Draw(t, label+" rkey?") {
		return uri, uri.String()
	}
	raw = genRecordKey().Draw(t, label+" rkey")
	rkey := RecordKey(raw)
	uri.recordKey = &rkey
	return uri, uri.String()
}

var dir = NewDirectory(http.DefaultClient, net.DefaultResolver, 5*time.Minute)

func TestParseRawURI(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		_, s := genURIHandle(t, "uri")
		t.Log(s)
		_, err := ParseRawURI(s)
		if err != nil {
			t.Fatal(err)
		}
	})
	rapid.Check(t, func(t *rapid.T) {
		u, s := genURI(t, "uri")
		t.Log(s)
		raw, err := ParseRawURI(s)
		if err != nil {
			t.Fatal(err)
		}
		uri, err := raw.URI(context.Background(), dir)
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
			_, err = uri.URI(context.Background(), dir)
			if err == nil {
				t.Errorf("expected error for %s", u)
			}
		}
	}
	for _, u := range invalidURI {
		fn(u)
	}
	for _, u := range invalidURILexicon {
		fn(u)
	}
	for _, u := range validURI {
		uri, err := ParseRawURI(u)
		if err != nil {
			t.Fatal(u, err)
		}
		if !testing.Short() {
			_, err := uri.URI(context.Background(), dir)
			if err != nil {
				t.Fatal(u, err)
			}
		}
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
