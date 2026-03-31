package atproto

import (
	"context"
	"net"
	"net/http"
	"testing"

	"pgregory.net/rapid"
)

var (
	invalidHandles = []string{
		"jo@hn.test",
		"💩.test",
		"john..test",
		"xn--bcher-.tld",
		"john.0",
		"cn.8",
		"www.masełkowski.pl.com",
		"org",
		"name.org.",
		// forbidden TLD
		"2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3zzui5du4xyclen53wid.onion",
		"laptop.local",
		"blah.arpa",
	}
	realHandles = []string{
		"anhgelus.world",
		"hailey.at",
		"mewsse.pet",
	}
)

func TestParseHandle(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		domain, _ := genDomain(t, false, "domain")
		h, err := ParseHandle(domain.Authority)
		if err != nil {
			t.Fatal(err)
		}
		if h != Handle(domain.Authority) {
			t.Errorf("invalid handle: %s, wanted %s", h, domain.Authority)
		}
	})
	for _, h := range invalidHandles {
		_, err := ParseHandle(h)
		if err == nil {
			t.Errorf("expected error for %s", h)
		}
	}
}

func TestDirectory_ResolveHandle(t *testing.T) {
	if testing.Short() {
		t.Skip("not doing http requests in short")
	}
	dir := NewDirectory(http.DefaultClient, net.DefaultResolver)
	for _, handle := range realHandles {
		h, err := ParseHandle(handle)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := dir.ResolveHandle(context.Background(), h)
		if err != nil {
			t.Fatal(err)
		}
		hh, ok := doc.Handle()
		if !ok {
			t.Fatal("impossible to get handle of", h)
		}
		if hh != h {
			t.Errorf("invalid also known as: %v, must contains %s", doc.AlsoKnownAs, "at://"+h.String())
		}
		t.Logf("%s's did: %s", h, doc.DID)
	}
}
