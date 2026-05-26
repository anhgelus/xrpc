package jetsream

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strconv"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

// Option used in the [Feed].
type Option struct {
	Collections         []*atproto.NSID
	DIDs                []*atproto.DID
	MaxMessageSizeBytes uint
	Cursor              uint64
}

// Feed is connected to a Jetstream.
type Feed struct {
	conn      *websocket.Conn
	closer    context.CancelFunc
	url       *url.URL
	ch        chan Event
	wait      chan struct{}
	connected atomic.Bool
	// Last Cursor sent by the server.
	Cursor uint64
}

// Connect to the [Feed] with the given [url.URL] and the given [Option].
func Connect(ctx context.Context, u *url.URL, log *slog.Logger, opt *Option) (*Feed, error) {
	if opt == nil {
		opt = &Option{}
	}
	u.Scheme = "wss"
	q := u.Query()
	for _, c := range opt.Collections {
		q.Add("wantedCollections", c.String())
	}
	for _, d := range opt.DIDs {
		q.Add("wantedDids", d.String())
	}
	if opt.MaxMessageSizeBytes > 0 {
		q.Add("maxMessageSizeBytes", strconv.Itoa(int(opt.MaxMessageSizeBytes)))
	}
	u.RawQuery = q.Encode()
	f := &Feed{url: u, Cursor: opt.Cursor}
	return f, f.Reconnect(ctx, log)
}

func (f *Feed) Disconnect(code websocket.StatusCode, msg string) error {
	f.connected.Store(false)
	f.closer()
	<-f.wait
	close(f.ch)
	return f.conn.Close(code, msg)
}

func (f *Feed) Reconnect(ctx context.Context, log *slog.Logger) error {
	if f.Connected() {
		log.Info("disconnecting...")
		err := f.Disconnect(websocket.StatusServiceRestart, "reconnecting")
		if err != nil {
			log.Error("disconnected", "error", err)
		} else {
			log.Info("disconnected")
		}
	}
	select {
	case <-ctx.Done():
		log.Warn("cannot restart", "reason", ctx.Err())
		return nil
	default:
	}
	if f.Cursor > 0 {
		q := f.url.Query()
		q.Set("cursor", strconv.Itoa(int(f.Cursor)))
		f.url.RawQuery = q.Encode()
	}
	target := f.url.String()
	log.Info("connecting...", "url", target)
	conn, _, err := websocket.Dial(ctx, target, nil)
	if err != nil {
		return err
	}
	log.Info("connected")
	f.conn = conn
	f.ch = make(chan Event, 1)
	f.wait = make(chan struct{})
	var ctxL context.Context
	ctxL, f.closer = context.WithCancel(ctx)
	go read(ctxL, log, f)
	f.connected.Store(true)
	return nil
}

func (f *Feed) Connected() bool {
	return f.connected.Load()
}

func (f *Feed) Listen() <-chan Event {
	return f.ch
}

func read(ctx context.Context, log *slog.Logger, f *Feed) {
	for {
		_, b, err := f.conn.Read(ctx)
		select {
		case <-ctx.Done():
			if !f.Connected() {
				close(f.wait)
				return
			}
			close(f.wait)
			log.Info("disconnecting: context finished")
			err = f.Disconnect(websocket.StatusNormalClosure, "good bye :3")
			if err != nil {
				log.Error("disconnected", "error", err)
			} else {
				log.Info("disconnected")
			}
			return
		default:
		}
		if err != nil {
			log.Error("reading websocket", "error", err)
			log.Warn("disconnecting...")
			close(f.wait)
			err = f.Disconnect(websocket.StatusProtocolError, "cannot read data")
			if err != nil {
				log.Error("disconnected", "error", err)
			} else {
				log.Info("disconnected")
			}
			return
		}
		var e Event
		err = json.Unmarshal(b, &e)
		if err != nil {
			log.Error("cannot unmarshal event", "error", err, "raw", b)
			continue
		}
		f.Cursor = e.TimeUs
		f.ch <- e
	}
}

type SubscriberSourcedMessage interface {
	json.Marshaler
	// Type of the message.
	Type() string
}

func marshalJSON[T SubscriberSourcedMessage](v T) ([]byte, error) {
	mp, err := xrpc.MarshalToMap(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"type": v.Type(), "payload": mp})
}

// SendMessage sends a [SubscriberSourcedMessage] to the websocket.
func (f *Feed) SendMessage(ctx context.Context, msg SubscriberSourcedMessage) error {
	return wsjson.Write(ctx, f.conn, msg)
}

type SubscriberOptionsUpdateMsg struct {
	Collections         []*atproto.NSID `json:"wantedCollections"`
	DIDs                []*atproto.DID  `json:"wantedDids"`
	MaxMessageSizeBytes uint            `json:"maxMessageSizeBytes"`
}

func (s *SubscriberOptionsUpdateMsg) Type() string {
	return "options_update"
}

func (s *SubscriberOptionsUpdateMsg) MarshalJSON() ([]byte, error) {
	return marshalJSON(s)
}
