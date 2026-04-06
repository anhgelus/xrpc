package server

import (
	"context"
	"encoding/json"
	"errors"

	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

var collection = atproto.NewNSIDBuilder("com.atproto.server")

// AccountInactiveReason is a known reason for an inactive account.
type AccountInactiveReason string

const (
	AccountTakedown    AccountInactiveReason = "takendown"
	AccountSuspended   AccountInactiveReason = "suspended"
	AccountDeactivated AccountInactiveReason = "deactivated"
)

// CreateSessionResult contains data returned from [CreateSession].
type CreateSessionResult struct {
	// Client is the new authentificated client.
	Client *xrpc.AuthClient
	// Handle of the user.
	Handle atproto.Handle `json:"handle"`
	// DID of the user.
	DID *atproto.DID `json:"did"`
	// DIDDocument is not required.
	DIDDocument *atproto.DIDDocument `json:"didDoc,omitempty"`
	// Email of the user.
	// It is not required.
	Email string `json:"email"`
	// EmailConfirmed is true if the user has confirmed their email.
	// It is not required.
	EmailConfirmed bool `json:"emailConfirmed"`
	// EmailAuthFactor is true if the user is using 2FA based on their [CreateSessionResult.Email].
	// It is not required.
	EmailAuthFactor bool `json:"emailAuthFactor"`
	// Active indicates whether the account is active.
	Active bool `json:"active"`
	// Status is the reason if [CreateSessionResult.Active] is false.
	// See [AccountInactiveReason].
	Status AccountInactiveReason `json:"status,omitempty"`
}

// CreateSessionRequest is used in [CreateSession] to authentificate a [xrpc.Client].
type CreateSessionRequest struct {
	// Identifier of the user.
	// Can be an [atproto.Handle] or an [atproto.DID].
	Identifier string `json:"identifier"`
	// Password of the user.
	Password string `json:"password"`
	// AuthFactorToken of the user.
	// Not required.
	AuthFactorToken string `json:"authFactorToken,omitempty"`
	// AllowTakenDown does not throw an error if the account is [AccountTakedown].
	AllowTakenDown bool `json:"allowTakenDown,omitempty"`
}

type sessionResult struct {
	CreateSessionResult
	AccessJWT  string `json:"accessJwt"`
	RefreshJWT string `json:"refreshJwt"`
}

// Standard errors returned by server lexicons.
var (
	ErrAccountTakedown         = xrpc.ErrStandard("AccountTakedown")
	ErrAuthFactorTokenRequired = xrpc.ErrStandard("AuthFactorTokenRequired")
)

// CreateSession using [xrpc.JWTAuth].
func CreateSession(
	ctx context.Context,
	client xrpc.Client,
	server string,
	params CreateSessionRequest,
) (*CreateSessionResult, error) {
	rb := client.NewRequest().
		Server(server).
		Endpoint(collection.Name("createSession").Build())
	b, err := client.Procedure(ctx, rb, xrpc.AsJsonBodyRequest(params))
	if err != nil {
		return nil, err
	}
	var resp sessionResult
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	resp.Client = xrpc.NewAuthClient(client, xrpc.NewJWTAuth(resp.DID), server)
	return &resp.CreateSessionResult, nil
}

var ErrNotJWT = errors.New("not JWT auth method")

// RefreshSession using [xrpc.JWTAuth].
//
// Returns [ErrNotJWT] if the [xrpc.AuthClient] is not using [xrpc.JWTAuth] as [xrpc.Auth] method.
func RefreshSession(ctx context.Context, client *xrpc.AuthClient) error {
	auth, ok := client.Auth().(*xrpc.JWTAuth)
	if !ok {
		return ErrNotJWT
	}
	rb := client.NewRequest().
		Auth(auth.AuthRequestRefresh()).
		Endpoint(collection.Name("refreshSession").Build())
	b, err := client.Procedure(ctx, rb, nil)
	if err != nil {
		return err
	}
	var resp sessionResult
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return err
	}
	auth.Refresh(resp.AccessJWT, resp.RefreshJWT)
	return nil
}

// DeleteSession (logout) using [xrpc.JWTAuth].
//
// Returns [ErrNotJWT] if the [xrpc.AuthClient] is not using [xrpc.JWTAuth] as [xrpc.Auth] method.
func DeleteSession(ctx context.Context, client *xrpc.AuthClient) error {
	_, ok := client.Auth().(*xrpc.JWTAuth)
	if !ok {
		return ErrNotJWT
	}
	rb := client.NewRequest().Endpoint(collection.Name("deleteSession").Build())
	_, err := client.Procedure(ctx, rb, nil)
	return err
}
