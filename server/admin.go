package server

import (
	"net/http"

	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

// AdminAuth is a basic [xrpc.Auth].
// It is used when an endpoint requires admin privileges.
// It must not be used with endpoint that requires to be linked with an account: [AdminAuth.DID] always panics.
//
// NOTE: [AdminAuth] could be moved into a future admin package.
type AdminAuth struct {
	password string
}

// NewAdminAuth creates a new [AdminAuth].
func NewAdminAuth(password string) *AdminAuth {
	return &AdminAuth{password}
}

func (a *AdminAuth) DID() *atproto.NSID {
	panic("cannot get DID from an admin auth")
}

func (a *AdminAuth) AuthRequest(req *http.Request) {
	req.SetBasicAuth("admin", a.password)
}

func (a *AdminAuth) IsInvalidAuth(err xrpc.ErrResponse) bool {
	return err.StatusCode == http.StatusUnauthorized
}
