package xrpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"anhgelus.world/xrpc/atproto"
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
	if !rb.useRelay {
		doc, err := dir.ResolveDID(ctx, did)
		if err != nil {
			return rb, err
		}
		var ok bool
		rb.server, ok = doc.PDS()
		if !ok {
			return rb, atproto.ErrCannotFindPDS
		}
	}
	return rb, nil
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

// Build returns a valid string representation of the request's endpoint and true if it must use CBOR.
//
// Panics if server or endpoint is not set.
func (rb RequestBuilder) Build(method string, body BodyRequestConverter) (*http.Request, bool, error) {
	if rb.server == "" {
		panic("cannot finish: server (PDS or relay) is not set")
	}
	if rb.endpoint == nil {
		panic("cannot finish: endpoint is not set")
	}
	var content io.Reader
	var b BodyRequest
	var err error
	if body != nil {
		b, err = body.AsBodyRequest(rb.useRelay)
		if err != nil {
			return nil, false, err
		}
		content = bytes.NewReader(b.Content)
	}

	req, err := http.NewRequest(method, rb.String(), content)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", rb.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", b.ContentType)
	}
	if rb.auth != nil {
		rb.auth.AuthRequest(req)
	}
	return req, rb.useRelay, nil
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

// BodyRequestConverter is a wrapper used to send data during a [Client.Procedure].
//
// See [EncodeBodyRequest] and [BodyRequest].
type BodyRequestConverter interface {
	// AsBodyRequest converts an [BodyRequestConverter] into a [BodyRequest].
	// It takes a bool set to true if it must encode the data into CBOR.
	AsBodyRequest(bool) (BodyRequest, error)
}

// EncodeBodyRequest is a [BodyRequest] that encodes the data.
//
// See [AsEncodeBodyRequest] to create a new [EncodeBodyRequest].
type EncodeBodyRequest struct {
	any
}

func (e EncodeBodyRequest) AsBodyRequest(useCbor bool) (BodyRequest, error) {
	b, err := Marshal(e, useCbor)
	return BodyRequest{b, "application/json"}, err
}

func AsEncodeBodyRequest(v any) EncodeBodyRequest {
	return EncodeBodyRequest{v}
}

// BodyRequest contains data sent during a [Client.Procedure].
//
// See [EncodeBodyRequest] to encode anything.
type BodyRequest struct {
	// Content of the request.
	Content []byte
	// Type of the [RawBodyRequest.Content].
	ContentType string
}

func (b BodyRequest) AsBodyRequest(bool) (BodyRequest, error) {
	return b, nil
}
