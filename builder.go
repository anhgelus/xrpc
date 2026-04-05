package xrpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// RequestBuilder is used to create the request endpoint.
//
// Always use [Client.NewRequest] to get the default [RequestBuilder] of the [Client].
type RequestBuilder struct {
	useRelay  bool
	server    string
	userAgent string
	endpoint  *atproto.NSID
	params    url.Values
	auth      Auth
}

func (rb RequestBuilder) Server(server string) RequestBuilder {
	rb.server = strings.TrimSuffix(server, "/")
	return rb
}

func (rb RequestBuilder) UseRelay() RequestBuilder {
	rb.useRelay = true
	return rb
}

func (rb RequestBuilder) PDS(ctx context.Context, dir atproto.Directory, did *atproto.DID) (RequestBuilder, error) {
	var err error
	if !rb.useRelay {
		rb.server, err = did.PDS(ctx, dir)
	}
	return rb, err
}

func (rb RequestBuilder) Endpoint(endpoint *atproto.NSID) RequestBuilder {
	rb.endpoint = endpoint
	return rb
}

func (rb RequestBuilder) Params(params url.Values) RequestBuilder {
	rb.params = params
	return rb
}

func (rb RequestBuilder) UserAgent(ua string) RequestBuilder {
	rb.userAgent = ua
	return rb
}

func (rb RequestBuilder) Auth(auth Auth) RequestBuilder {
	rb.auth = auth
	return rb
}

func (rb RequestBuilder) GetAuth() Auth {
	return rb.auth
}

// Build returns a valid string representation of the request's endpoint.
//
// Panics if server or endpoint is not set.
func (rb RequestBuilder) Build(method string, body BodyRequest) (*http.Request, error) {
	if rb.server == "" {
		panic("cannot finish: server (PDS or relay) is not set")
	}
	if rb.endpoint == nil {
		panic("cannot finish: endpoint is not set")
	}
	var content io.Reader
	if body != nil {
		b, err := body.Body()
		if err != nil {
			return nil, err
		}
		content = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, rb.String(), content)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", rb.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", body.ContentType())
	}
	if rb.auth != nil {
		rb.auth.AuthRequest(req)
	}
	return req, nil
}

func (rb RequestBuilder) String() string {
	var sb strings.Builder
	sb.WriteString(rb.server)
	sb.WriteRune('/')
	sb.WriteString(BaseURL)
	if rb.endpoint != nil {
		sb.WriteRune('/')
		sb.WriteString(rb.endpoint.String())
	}
	if rb.params != nil {
		sb.WriteRune('?')
		sb.WriteString(rb.params.Encode())
	}
	return sb.String()
}

// BodyRequest contains data sent during a [Client.Procedure].
//
// See [JsonBodyRequest] to encode anything into JSON.
// See [RawBodyRequest] to send raw data.
type BodyRequest interface {
	// Body returns the body of the Request.
	Body() ([]byte, error)
	// ContentType return the Content-Type header of the [BodyRequest.Body].
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
	return Marshal(j)
}

func (j JsonBodyRequest) ContentType() string {
	return "application/json"
}

// RawBodyRequest is a [BodyRequest] that contains raw data.
type RawBodyRequest struct {
	// Content of the request.
	Content []byte
	// Type of the [RawBodyRequest.Content].
	Type string
}

func (r RawBodyRequest) Body() ([]byte, error) {
	return r.Content, nil
}

func (r RawBodyRequest) ContentType() string {
	return r.Type
}
