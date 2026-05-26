package jetsream

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strconv"
	"sync/atomic"

	"anhgelus.world/xrpc"
	"anhgelus.world/xrpc/atproto"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Options used in the [Feed].
type Options struct {
	Collections         []*atproto.NSID
	DIDs                []*atproto.DID
	MaxMessageSizeBytes uint
	Cursor              uint64
}

// Feed connected to a Jetstream.
//
// Can be used concurrently.
//
// See [New], [Feed.Connect] and [Feed.Listen].
// See [Options].
type Feed struct {
	log       *slog.Logger
	dialOpt   *websocket.DialOptions
	conn      *websocket.Conn
	closer    context.CancelFunc
	url       *url.URL
	ch        chan Event
	wait      chan struct{}
	connected atomic.Bool
	// Last Cursor sent by the server.
	Cursor uint64
}

// New creates a [Feed].
//
// [Options] can be nil.
func New(log *slog.Logger, u *url.URL, opt *Options) (*Feed, error) {
	if opt == nil {
		opt = &Options{}
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
	f := &Feed{log: log, url: u, Cursor: opt.Cursor}
	return f, nil
}

// Connect the [Feed] to the Jetstream.
//
// If the [Feed] is already [Feed.Connected], it reconnects it with the given options.
//
// [websocket.DialOptions] can be nil.
func (f *Feed) Connect(ctx context.Context, websocketOpt *websocket.DialOptions) error {
	f.dialOpt = websocketOpt
	return f.Reconnect(ctx)
}

// Disconnect the [Feed] from the Jetstream.
//
// Multiple calls are no-op.
func (f *Feed) Disconnect(code websocket.StatusCode, msg string) error {
	if !f.Connected() {
		return nil
	}
	f.connected.Store(false)
	f.closer()
	<-f.wait
	close(f.ch)
	return f.conn.Close(code, msg)
}

// Reconnect the [Feed] to the Jetstream.
//
// Use [Connect] if you to change the [websocket.DialOptions] used.
func (f *Feed) Reconnect(ctx context.Context) error {
	if f.Connected() {
		f.log.Info("disconnecting and reconnecting...")
		err := f.Disconnect(websocket.StatusServiceRestart, "reconnecting")
		if err != nil {
			f.log.Error("disconnected", "error", err)
		} else {
			f.log.Info("disconnected")
		}
	}
	if f.Cursor > 0 {
		q := f.url.Query()
		q.Set("cursor", strconv.Itoa(int(f.Cursor)))
		f.url.RawQuery = q.Encode()
	}
	f.log.Info("connecting...", "url", f.url.Redacted())
	select {
	case <-ctx.Done():
		f.log.Warn("cannot restart", "reason", ctx.Err())
		return nil
	default:
	}
	conn, _, err := websocket.Dial(ctx, f.url.String(), nil)
	if err != nil {
		return err
	}
	f.log.Info("connected")
	f.conn = conn
	f.ch = make(chan Event, 1)
	f.wait = make(chan struct{})
	var ctxL context.Context
	ctxL, f.closer = context.WithCancel(ctx)
	go f.read(ctxL)
	f.connected.Store(true)
	return nil
}

// Connected returns true if the [Feed] is connected.
func (f *Feed) Connected() bool {
	return f.connected.Load()
}

// Listen returns the channel receiving [Event].
func (f *Feed) Listen() <-chan Event {
	return f.ch
}

func (f *Feed) read(ctx context.Context) {
	for {
		_, b, err := f.conn.Read(ctx)
		select {
		case <-ctx.Done():
			if !f.Connected() {
				close(f.wait)
				return
			}
			close(f.wait)
			f.log.Info("disconnecting: context finished")
			err = f.Disconnect(websocket.StatusNormalClosure, "good bye :3")
			if err != nil {
				f.log.Error("disconnected", "error", err)
			} else {
				f.log.Info("disconnected")
			}
			return
		default:
		}
		if err != nil {
			f.log.Error("reading websocket", "error", err)
			f.log.Warn("disconnecting...")
			close(f.wait)
			err = f.Disconnect(websocket.StatusProtocolError, "cannot read data")
			if err != nil {
				f.log.Error("disconnected", "error", err)
			} else {
				f.log.Info("disconnected")
			}
			return
		}
		var e Event
		err = json.Unmarshal(b, &e)
		if err != nil {
			f.log.Error("cannot unmarshal event", "error", err, "raw", b)
			continue
		}
		f.Cursor = e.TimeUs
		f.ch <- e
	}
}

// SubscriberSourcedMessage is a message sent by the client (us).
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

// SubscriberOptionsUpdateMsg is used to update the [Options] of the [Feed].
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
