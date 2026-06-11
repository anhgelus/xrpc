package jetsream

import (
	"context"
	"log/slog"
	"math/rand"
	"net/url"
	"testing"
	"time"

	_ "pgregory.net/rapid"
)

var jetstream = [4]string{
	"wss://jetstream2.us-east.bsky.network/subscribe",
	"wss://jetstream1.us-east.bsky.network/subscribe",
	"wss://jetstream2.fr.hose.cam/subscribe",
	"wss://jetstream.fire.hose.cam/subscribe"}

func genBase(ctx context.Context, log *slog.Logger) (*Feed, error) {
	v := jetstream[rand.Intn(len(jetstream))]
	u, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	f, err := New(log, u, nil)
	if err != nil {
		return nil, err
	}
	return f, f.Connect(ctx, nil)
}

func TestFeed_Connect(t *testing.T) {
	if testing.Short() {
		t.Skip("not doing requests in short mode")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	log := slog.Default()
	f, err := genBase(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	var lastCursor uint64
	data := make(map[string]uint, 3)
	i := 0
	for e := range f.Listen() {
		i++
		lastCursor = e.TimeUs
		data[string(e.Kind)]++
		if i%500 == 0 {
			log.Info(
				"logs got",
				CommitKind, data[string(CommitKind)],
				AccountKind, data[string(AccountKind)],
				IdentityKind, data[string(IdentityKind)],
			)
		}
	}
	if f.Connected() {
		t.Error("connected after listening closed")
	}
	if lastCursor != f.Cursor {
		t.Errorf("invalid cursor: %d, wanted %d", f.Cursor, lastCursor)
	}
}

func TestFeed_Reconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("not doing requests in short mode")
	}
	f, err := genBase(t.Context(), slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	for i/1000 < 5 {
		<-f.Listen()
		i++
		if i%(1000+rand.Intn(100)) == 0 {
			err = f.Reconnect(t.Context())
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
