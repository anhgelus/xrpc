package xrpc

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// Auth describes how to authentificate a request.
//
// Can be safely used concurrently.
//
// See [NewJWTAuth] to use [JWTAuth] to authentificate requests.
type Auth interface {
	DID() *atproto.DID
	// AuthRequest authentificates the given [http.Request].
	AuthRequest(*http.Request)
	// IsInvalidAuth returns true if the [ErrResponse] indicates an invalid auth.
	IsInvalidAuth(err ErrResponse) bool
}

// JWTAuth contains [Auth] data using JWT tokens.
type JWTAuth struct {
	mu      sync.RWMutex
	access  string
	refresh string
	did     *atproto.DID
}

// NewJWTAuth creates a new [JWTAuth].
//
// You must call [JWTAuth.Refresh] before using it.
func NewJWTAuth(did *atproto.DID) *JWTAuth {
	return &JWTAuth{did: did}
}

func (a *JWTAuth) AccessToken() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.access
}

func (a *JWTAuth) RefreshToken() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.refresh
}

func (a *JWTAuth) DID() *atproto.DID {
	return a.did
}

func (a *JWTAuth) AuthRequest(req *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req.Header.Add("Authorization", "Bearer "+a.access)
}

// Refresh the [JWTAuth] with provided tokens.
func (a *JWTAuth) Refresh(access, refresh string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.access = access
	a.refresh = refresh
}

// AuthRequestRefresh returns an [Auth] using refresh token instead of access token.
func (a *JWTAuth) AuthRequestRefresh() Auth {
	return &JWTAuth{access: a.refresh}
}

const InvalidJWTAuth = "ExpiredToken"

func (a *JWTAuth) IsInvalidAuth(err ErrResponse) bool {
	if err.StatusCode != http.StatusBadRequest {
		return false
	}
	var r ErrStandardResponse
	if !errors.As(err, &r) {
		return false
	}
	return r.ErrorKind == InvalidJWTAuth
}

// AuthClient is a [Client] used if an endpoint requires authentification.
type AuthClient struct {
	Client
	server string
	auth   Auth
}

// NewAuthClient creates a new [AuthClient].
//
// Prefer using [server.CreateSession].
//
// See [NewAuthClientFetchServer] if you don't have the server linked with the [Auth].
func NewAuthClient(base Client, auth Auth, server string) *AuthClient {
	return &AuthClient{base, server, auth}
}

// NewAuthClientFetchServer creates a new [AuthClient] and fetch the server linked with the [Auth].
//
// Prefer using [server.CreateSession].
//
// See [NewAuthClient] if you already have a server.
func NewAuthClientFetchServer(ctx context.Context, base Client, auth Auth) (*AuthClient, error) {
	pds, err := auth.DID().PDS(ctx, base.Directory())
	return NewAuthClient(base, auth, pds), err
}

func (c *AuthClient) NewRequest() RequestBuilder {
	return c.Client.
		NewRequest().
		Server(c.server).
		Auth(c.auth)
}

func (c *AuthClient) Auth() Auth {
	return c.auth
}
