package xrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

const (
	Query     = http.MethodGet
	Procedure = http.MethodPost
	BaseURL   = "xrpc"
)

// Client is a simple ATProto XRPC client.
//
// Can be used concurrently by multiple goroutines.
//
// See [NewClient] to create a new [Client].
type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{client}
}

// Query performs an XRPC [Query].
//
// pds is the base url of the PDS.
// id is the [atproto.NSID] of the endpoint used.
// params are the [url.Values] used during the call.
//
// Returns [ErrRequest] if the response code indicates an error >= 400.
func (c *Client) Query(ctx context.Context, pds string, id *atproto.NSID, params url.Values) ([]byte, error) {
	return c.do(ctx, Query, pds, id, params, nil)
}

// BodyRequest contains data sent during a [Client.Procedure].
//
// See [JsonBodyRequest] to encode anything into JSON.
// See [RawBodyRequest] to send raw data.
type BodyRequest interface {
	Body() ([]byte, error)
	ContentType() string
}

// JsonBodyRequest is a [BodyRequest] that encodes the data into JSON.
//
// See [AsJsonBodyRequest] to create a new [JsonBodyRequest].
type JsonBodyRequest struct {
	any
}

func AsJsonBodyRequest(v any) JsonBodyRequest {
	return JsonBodyRequest{v}
}

func (j JsonBodyRequest) Body() ([]byte, error) {
	return json.Marshal(j)
}

func (j JsonBodyRequest) ContentType() string {
	return "application/json"
}

// RawBodyRequest is a [BodyRequest] that contains raw data.
type RawBodyRequest struct {
	Content []byte
	Type    string
}

func (r RawBodyRequest) Body() ([]byte, error) {
	return r.Content, nil
}

func (r RawBodyRequest) ContentType() string {
	return r.Type
}

// Procedure performs an XRPC [Procedure].
//
// pds is the base url of the PDS.
// id is the [atproto.NSID] of the endpoint used.
// params are the [url.Values] used during the call.
// body is the body of the request.
//
// Returns [ErrRequest] if the response code indicates an error >= 400.
func (c *Client) Procedure(ctx context.Context, pds string, id *atproto.NSID, params url.Values, body BodyRequest) ([]byte, error) {
	return c.do(ctx, Procedure, pds, id, params, body)
}

func (c *Client) do(ctx context.Context, method string, pds string, id *atproto.NSID, params url.Values, body BodyRequest) ([]byte, error) {
	var content io.Reader
	if body != nil {
		b, err := body.Body()
		if err != nil {
			return nil, err
		}
		content = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, pds, content)
	if err != nil {
		return nil, err
	}
	req.URL.Path += BaseURL + "/" + id.String() + "?" + params.Encode()
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", body.ContentType())
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, ErrRequest{resp.StatusCode, b}
	}
	return io.ReadAll(resp.Body)
}

type ErrRequest struct {
	statusCode int
	content    []byte
}

func (r ErrRequest) Error() string {
	if r.content != nil {
		var v struct {
			Error   string `json:"string"`
			Message string `json:"omitempty"`
		}
		err := json.Unmarshal(r.content, &v)
		if err != nil {
			return fmt.Sprintf("%s (status code: %d)", r.content, r.statusCode)
		}
		if v.Message != "" {
			return fmt.Sprintf("%s: %s", v.Error, v.Message)
		}
		return v.Error
	}
	return fmt.Sprintf("invalid status code: %d", r.statusCode)
}
