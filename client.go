package xrpc

import (
	"context"
	"encoding/json"
	"errors"
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
// See [NewAuthClient] to create a new [AuthClient].
// See [NewCompatClient] to create a new [CompatClient].
type Client interface {
	// Query performs an XRPC [Query].
	//
	// Returns [ErrResponse] if the response code indicates an error >= 400.
	Query(context.Context, RequestBuilder) ([]byte, error)
	// Procedure performs an XRPC [Procedure].
	//
	// Returns [ErrResponse] if the response code indicates an error >= 400.
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
	Directory() atproto.Directory
}

var ErrInvalidAuth = errors.New("invalid auth")

// BaseClient is a simple ATProto XRPC client.
type BaseClient struct {
	client *http.Client
	dir    atproto.Directory
}

// NewClient creates a new [BaseClient].
func NewClient(client *http.Client, dir atproto.Directory) *BaseClient {
	return &BaseClient{client, dir}
}

func (c *BaseClient) HTTP() *http.Client {
	return c.client
}

func (c *BaseClient) Directory() atproto.Directory {
	return c.dir
}

func (c *BaseClient) Query(ctx context.Context, rb RequestBuilder) ([]byte, error) {
	return c.do(ctx, Query, rb, nil)
}

func (c *BaseClient) Procedure(ctx context.Context, rb RequestBuilder, body BodyRequest) ([]byte, error) {
	return c.do(ctx, Procedure, rb, body)
}

func (c *BaseClient) do(ctx context.Context, method string, rb RequestBuilder, body BodyRequest) ([]byte, error) {
	req, err := rb.Build(method, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err := ErrResponse{resp.StatusCode, b}
		auth := rb.GetAuth()
		if auth == nil {
			return nil, err
		}
		if auth.IsInvalidAuth(err) {
			return nil, ErrInvalidAuth
		}
		return nil, err
	}
	return b, err
}

func (c *BaseClient) NewRequest() RequestBuilder {
	return RequestBuilder{}
}

type ErrStandard string

func (e ErrStandard) Error() string {
	return "defined lexicon: " + string(e)
}

// ErrStandardResponse represents a standard error response from the server following lexicon definition.
//
// You can check easily its lexicon type with [ErrStandard] and [errors.Is]:
//
//	var err ErrStandardResponse
//	errors.Is(err, xrpc.ErrRecordNotFound)
//
// Obtained from [ErrResponse].
type ErrStandardResponse struct {
	ErrorKind string `json:"error"`
	Message   string `json:"message,omitempty"`
}

func (r ErrStandardResponse) Error() string {
	if r.Message != "" {
		return fmt.Sprintf("%s: %s", r.ErrorKind, r.Message)
	}
	return r.ErrorKind
}

func (r ErrStandardResponse) Is(err error) bool {
	std, ok := err.(ErrStandard)
	if !ok {
		e, ok := err.(ErrStandardResponse)
		return ok && e.ErrorKind == r.ErrorKind
	}
	return r.ErrorKind == string(std)
}

// ErrResponse is returned by a [Client] when an [http.Response] contains a status code >= 400.
//
// Use [errors.As] can be used to convert an [ErrResponse] into an [ErrStandardResponse].
type ErrResponse struct {
	// StatusCode of the response.
	StatusCode int
	// Content of the response.
	Content []byte
}

func (r ErrResponse) As(target any) bool {
	v, ok := target.(*ErrStandardResponse)
	if !ok || r.Content == nil {
		return false
	}
	err := json.Unmarshal(r.Content, v)
	if err != nil {
		return false
	}
	return len(v.ErrorKind) > 0
}

func (r ErrResponse) Error() string {
	if r.Content != nil {
		var std ErrStandardResponse
		if errors.As(r, &std) {
			return std.Error()
		}
		return fmt.Sprintf("%s (status code: %d)", r.Content, r.StatusCode)
	}
	return fmt.Sprintf("invalid status code: %d", r.StatusCode)
}
