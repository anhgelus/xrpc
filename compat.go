package xrpc

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// CompatClient is a [Client] that is compatible with Indigo's LexClient.
//
// It has the same properties of the inner [Client], except that it only works for a certain server while used by
// Indigo.
type CompatClient struct {
	Client
	server string
}

// NewCompatClient creates a new [CompatClient].
func NewCompatClient(base Client, server string) *CompatClient {
	return &CompatClient{base, server}
}

func (c *CompatClient) NewRequest() RequestBuilder {
	return c.Client.NewRequest().Server(c.server)
}

func (c *CompatClient) LexDo(
	ctx context.Context,
	method string,
	inputEncoding string,
	endpoint string,
	params map[string]any,
	bodyData any,
	out any,
) error {
	nsid, err := atproto.ParseNSID(endpoint)
	if err != nil {
		return err
	}
	val, err := parseIndigoParams(params)
	if err != nil {
		return err
	}
	req := c.NewRequest().Endpoint(nsid).Params(val)

	var b []byte
	switch method {
	case Query:
		b, err = c.Query(ctx, req)
	case Procedure:
		var body RawBodyRequest
		body.Type = inputEncoding
		if r, ok := bodyData.(io.Reader); ok {
			body.Content, err = io.ReadAll(r)
		} else {
			body.Content, err = json.Marshal(bodyData)
			if body.Type == "" {
				body.Type = "application/json"
			}
		}
		if err != nil {
			return err
		}
		b, err = c.Procedure(ctx, req, body)
	}
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(b, &out)
}

func parseIndigoParams(params map[string]any) (url.Values, error) {
	out := make(url.Values)
	for k, v := range params {
		if v == nil {
			out.Add(k, "")
			continue
		}
		if s, ok := v.(encoding.TextMarshaler); ok {
			out.Add(k, fmt.Sprint(s))
			continue
		}
		ref := reflect.ValueOf(v)
		switch ref.Kind() {
		case reflect.Int, reflect.Bool, reflect.Uint, reflect.Float32, reflect.Float64, reflect.String:
			out.Add(k, fmt.Sprint())
		case reflect.Slice:
			for i := range ref.Len() {
				elem := ref.Index(i)
				if elem.IsNil() {
					out.Add(k, "")
				} else if s, ok := v.(encoding.TextMarshaler); ok {
					out.Add(k, fmt.Sprint(s))
				} else {
					switch elem.Kind() {
					case reflect.Int, reflect.Bool, reflect.Uint, reflect.Float32, reflect.Float64, reflect.String:
						out.Add(k, fmt.Sprint())
					default:
						return nil, fmt.Errorf("can't marshal query param '%s' with type: %T", k, v)
					}
				}
			}
		default:
			return nil, fmt.Errorf("can't marshal query param '%s' with type: %T", k, v)
		}
	}
	return out, nil
}
