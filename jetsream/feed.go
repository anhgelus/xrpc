package jetsream

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strconv"

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
	Cursor              uint
}

// Feed is connected to a Jetstream.
type Feed struct {
	conn   *websocket.Conn
	closer context.CancelFunc
	url    string
	// Last Cursor sent by the server.
	Cursor   uint
	Listener <-chan Event
	sender   chan<- Event
}

// Connect to the [Feed] with the given [url.URL] and the given [Option].
func Connect(ctx context.Context, u *url.URL, log *slog.Logger, opt *Option) (*Feed, error) {
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
	if opt.Cursor > 0 {
		q.Add("cursor", strconv.Itoa(int(opt.Cursor)))
	}
	target := u.String()
	f := &Feed{url: target}
	return f, f.Reconnect(ctx, log)
}

func (f *Feed) Disconnect(ctx context.Context, code websocket.StatusCode, msg string) error {
	f.closer()
	close(f.sender)
	f.Listener = nil
	return f.conn.Close(code, msg)
}

func (f *Feed) Reconnect(ctx context.Context, log *slog.Logger) error {
	if f.Listener != nil {
		log.Info("disconnecting...")
		err := f.Disconnect(ctx, websocket.StatusServiceRestart, "reconnecting")
		if err != nil {
			log.Error("disconnected", "error", err)
		} else {
			log.Info("disconnected")
		}
	}
	log.Info("connecting...", "url", f.url)
	conn, _, err := websocket.Dial(ctx, f.url, nil)
	if err != nil {
		return err
	}
	log.Info("connected")
	l := make(chan Event, 1)
	f.conn = conn
	f.Listener = l
	f.sender = l
	var ctxL context.Context
	ctxL, f.closer = context.WithCancel(ctx)
	go read(ctxL, log, f)
	return nil
}

func read(ctx context.Context, log *slog.Logger, f *Feed) {
	for {
		select {
		case <-ctx.Done():
			log.Info("exiting: context finished")
			return
		default:
			_, b, err := f.conn.Read(ctx)
			if err != nil {
				log.Error("reading websocket", "error", err)
				log.Warn("disconnecting...")
				err = f.Disconnect(ctx, websocket.StatusProtocolError, "cannot read data")
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
			f.Cursor = uint(e.TimeUs)
			f.sender <- e
		}
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
