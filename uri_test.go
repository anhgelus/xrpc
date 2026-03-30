package xrpc

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

var client Client

func getClient() Client {
	if client == nil {
		client = NewClient(http.DefaultClient, atproto.NewDirectory(http.DefaultClient, net.DefaultResolver, 5*time.Minute))
	}
	return client
}

var validURI = []string{
	"at://did:plc:zcanytzlaumjwgaopolw6wes/site.standard.document/3mhmdp3qobs2o",
	"at://did:plc:revjuqmkvrw6fnkxppqtszpv/site.standard.document/3mbfqhezge25u",
	"at://did:plc:vtqucb4iga7b5wzza3zbz4so/app.bsky.feed.post/3mhsdpqccys2i",
	"at://did:plc:vtqucb4iga7b5wzza3zbz4so/sh.tangled.repo/3mi4oatla6z22",
}

func TestClient_FetchURI(t *testing.T) {
	if testing.Short() {
		t.Skip("not doing http requests in short")
	}
	c := getClient()
	for _, raw := range validURI {
		uri, err := atproto.ParseURI(context.Background(), c.Directory(), raw)
		if err != nil {
			t.Fatal(err)
		}
		union, err := c.FetchURI(context.Background(), uri)
		if err != nil {
			t.Fatal(err)
		}
		if !union.Value.Collection().Is(uri.Collection()) {
			t.Errorf("invalid type: %s, wanted %s", union.Value.Collection(), uri.Collection())
		}
	}
}
