package jetsream

import (
	"context"
	"log/slog"
	"net/url"
	"testing"
	"time"

	_ "pgregory.net/rapid"
)

const jetstream = "wss://jetstream2.us-east.bsky.network/subscribe"

func genBase(ctx context.Context, log *slog.Logger) (*Feed, error) {
	u, err := url.Parse(jetstream)
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
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	log := slog.Default()
	f, err := genBase(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-f.Listen():
			i++
			if i%1030 == 0 {
				err = f.Reconnect(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
