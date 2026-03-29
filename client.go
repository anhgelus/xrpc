package xrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

const (
	Query     = http.MethodGet
	Procedure = http.MethodPost
	BaseURL   = "xrpc"
)

// Client represents a general ATProto XRPC client.
//
// Can be used concurrently by multiple goroutines.
//
// See [NewClient] to create a new [BaseClient].
type Client interface {
	// Query performs an XRPC [Query].
	//
	// Returns [ErrRequest] if the response code indicates an error >= 400.
	Query(context.Context, RequestBuilder) ([]byte, error)
	// Procedure performs an XRPC [Procedure].
	//
	// Returns [ErrRequest] if the response code indicates an error >= 400.
	Procedure(context.Context, RequestBuilder, BodyRequest) ([]byte, error)
	// FetchRawURI returns the [Record] pointed by the [atproto.RawURI].
	//
	// Returns [ErrIncompleteURI] if the [atproto.RawURI] doesn't contain enough information to get [Record].
	FetchURI(ctx context.Context, uri atproto.URI) (RecordStored[*Union], error)
	// NewRequest returns the base [RequestBuilder] used.
	NewRequest() RequestBuilder
	// HTTP returns the [http.Client] used by the [Client].
	HTTP() *http.Client
	// Directory returns the [atproto.Directory] used by the [Client].
	Directory() *atproto.Directory
}

// BaseClient is a simple ATProto XRPC client.
type BaseClient struct {
	client *http.Client
	dir    *atproto.Directory
}

// NewClient creates a new [BaseClient].
func NewClient(client *http.Client, dir *atproto.Directory) *BaseClient {
	return &BaseClient{client, dir}
}

func (c *BaseClient) HTTP() *http.Client {
	return c.client
}

func (c *BaseClient) Directory() *atproto.Directory {
	return c.dir
}

func (c *BaseClient) Query(ctx context.Context, rb RequestBuilder) ([]byte, error) {
	return c.do(ctx, Query, rb, nil)
}

func (c *BaseClient) Procedure(ctx context.Context, rb RequestBuilder, body BodyRequest) ([]byte, error) {
	return c.do(ctx, Procedure, rb, body)
}

func (c *BaseClient) do(ctx context.Context, method string, rb RequestBuilder, body BodyRequest) ([]byte, error) {
	var content io.Reader
	if body != nil {
		b, err := body.Body()
		if err != nil {
			return nil, err
		}
		content = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, rb.Build(), content)
	if err != nil {
		return nil, err
	}
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

func (c *BaseClient) NewRequest() RequestBuilder {
	return RequestBuilder{}
}

// ErrRequest is returned by a [Client] when an [http.Response] contains a status code >= 400.
type ErrRequest struct {
	// StatusCode of the response.
	StatusCode int
	// Content of the response.
	Content []byte
}

func (r ErrRequest) Error() string {
	if r.Content != nil {
		var v struct {
			Error   string `json:"string"`
			Message string `json:"omitempty"`
		}
		err := json.Unmarshal(r.Content, &v)
		if err != nil {
			return fmt.Sprintf("%s (status code: %d)", r.Content, r.StatusCode)
		}
		if v.Message != "" {
			return fmt.Sprintf("%s: %s", v.Error, v.Message)
		}
		return v.Error
	}
	return fmt.Sprintf("invalid status code: %d", r.StatusCode)
}
