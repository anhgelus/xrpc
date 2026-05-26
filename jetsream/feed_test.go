package jetsream

import (
	"context"
	"log/slog"
	"net/url"
	"testing"
	"time"
)

const jetstream = "wss://jetstream2.us-east.bsky.network/subscribe"

func genBase(ctx context.Context, log *slog.Logger) (*Feed, error) {
	u, err := url.Parse(jetstream)
	if err != nil {
		return nil, err
	}
	f, err := Connect(ctx, u, log, nil)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func TestFeed_Connect(t *testing.T) {
	if testing.Short() {
		t.Skip("not doing requests in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	log := slog.Default()
	f, err := genBase(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	var lastCursor uint64
	for e := range f.Listen() {
		lastCursor = e.TimeUs
		log.Info("event received!", "event", e)
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
			if i%500 == 0 {
				err = f.Reconnect(ctx, log)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
