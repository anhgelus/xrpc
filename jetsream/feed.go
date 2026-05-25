package jetsream

import (
	"context"
	"encoding/json"
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
	conn *websocket.Conn
	// Last Cursor sent by the server.
	Cursor uint
}

// Connect to the [Feed] with the given [url.URL] and the given [Option].
func Connect(ctx context.Context, u *url.URL, opt *Option) (*Feed, error) {
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
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}
	return &Feed{conn: conn}, nil
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
