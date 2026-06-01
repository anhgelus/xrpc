package xrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"anhgelus.world/xrpc/atproto"
)

// Standard values used by a XRPC client
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
// See [NewRelayClient] to create a new [RelayClient].
type Client interface {
	// Query performs an XRPC [Query].
	// Returns the content of the response and true if it is encoded with CBOR.
	//
	// Returns [ErrResponse] if the response code indicates an error >= 400.
	Query(context.Context, RequestBuilder) ([]byte, bool, error)
	// Procedure performs an XRPC [Procedure].
	// Returns the content of the response and true if it is encoded with CBOR.
	//
	// Returns [ErrResponse] if the response code indicates an error >= 400.
	Procedure(context.Context, RequestBuilder, BodyRequestConverter) ([]byte, bool, error)
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

// ErrInvalidAuth is returned when the [Auth] used is invalid.
type ErrInvalidAuth struct {
	inner error
}

func (err ErrInvalidAuth) Error() string {
	return "invalid auth"
}

func (err ErrInvalidAuth) Unwrap() error {
	return err.inner
}

// BaseClient is a simple ATProto XRPC client.
type BaseClient struct {
	UserAgent string
	client    *http.Client
	dir       atproto.Directory
}

// NewClient creates a new [BaseClient].
func NewClient(client *http.Client, dir atproto.Directory, userAgent string) *BaseClient {
	return &BaseClient{userAgent, client, dir}
}

func (c *BaseClient) HTTP() *http.Client {
	return c.client
}

func (c *BaseClient) Directory() atproto.Directory {
	return c.dir
}

func (c *BaseClient) Query(ctx context.Context, rb RequestBuilder) ([]byte, bool, error) {
	return c.do(ctx, Query, rb, nil)
}

func (c *BaseClient) Procedure(ctx context.Context, rb RequestBuilder, body BodyRequestConverter) ([]byte, bool, error) {
	return c.do(ctx, Procedure, rb, body)
}

func (c *BaseClient) do(ctx context.Context, method string, rb RequestBuilder, body BodyRequestConverter) ([]byte, bool, error) {
	req, useCbor, err := rb.Build(method, body)
	if err != nil {
		return nil, useCbor, err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, useCbor, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err := ErrResponse{resp.StatusCode, b}
		auth := rb.GetAuth()
		if auth == nil {
			return nil, useCbor, err
		}
		if auth.IsInvalidAuth(err) {
			return nil, useCbor, ErrInvalidAuth{err}
		}
		return nil, useCbor, err
	}
	return b, useCbor, err
}

func (c *BaseClient) NewRequest() RequestBuilder {
	return RequestBuilder{}
}

// ErrStandard is an error defined in the lexicon.
//
// Obtained from [ErrStandardResponse].
type ErrStandard string

func (e ErrStandard) Error() string {
	return "defined lexicon error: " + string(e)
}

// ErrStandardResponse represents a standard error response from the server following lexicon definition.
//
// You can check easily its lexicon type with [ErrStandard] and [errors.Is]:
//
//	var err ErrStandardResponse
//	errors.Is(err, xrpc.ErrRecordNotFound)
//
// You can get the [ErrStandard] with [errors.Unwrap] or with [errors.AsType].
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

func (r ErrStandardResponse) Unwrap() error {
	return ErrStandard(r.ErrorKind)
}

func (r ErrStandardResponse) Is(err error) bool {
	switch v := err.(type) {
	case ErrStandard:
		return r.ErrorKind == string(v)
	case ErrStandardResponse:
		return r.ErrorKind == v.ErrorKind
	default:
		return false
	}
}

func (r ErrStandardResponse) As(target any) bool {
	switch v := target.(type) {
	case *ErrStandard:
		*v = ErrStandard(r.ErrorKind)
		return true
	default:
		return false
	}
}

// ErrResponse is returned by a [Client] when an [http.Response] contains a status code >= 400.
//
// [errors.AsType] can be used to convert an [ErrResponse] into an [ErrStandardResponse].
type ErrResponse struct {
	// StatusCode of the response.
	StatusCode int
	// Content of the response.
	Content []byte
}

func (r ErrResponse) As(target any) bool {
	switch v := target.(type) {
	case *ErrStandardResponse:
		if r.Content == nil {
			return false
		}
		err := json.Unmarshal(r.Content, v)
		if err != nil {
			return false
		}
		return len(v.ErrorKind) > 0
	case *ErrStandard:
		var t ErrStandardResponse
		return errors.As(r, &t) && errors.As(t, v)
	default:
		return false
	}
}

func (r ErrResponse) Is(err error) bool {
	switch v := err.(type) {
	case ErrResponse:
		return r.StatusCode == v.StatusCode
	case ErrStandard:
		var conv ErrStandard
		return errors.As(r, &conv) && errors.Is(conv, err)
	default:
		return false
	}
}

func (r ErrResponse) Error() string {
	if r.Content != nil {
		if std, ok := errors.AsType[ErrStandardResponse](r); ok {
			return std.Error()
		}
		return fmt.Sprintf("%s (status code: %d)", r.Content, r.StatusCode)
	}
	return fmt.Sprintf("invalid status code: %d", r.StatusCode)
}

// RelayClient is a [Client] that communicates with a relay instead of a PDS.
type RelayClient struct {
	Client
	relay string
}

// NewRelayClient creates a new [RelayClient].
func NewRelayClient(client Client, relay string) *RelayClient {
	return &RelayClient{client, relay}
}

func (c *RelayClient) NewRequest() RequestBuilder {
	return c.Client.NewRequest().Server(c.relay).UseRelay()
}
