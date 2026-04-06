package atproto

import (
	"errors"
	"fmt"
	"net"
	"net/http"
)

// IsErrCannotParse returns true if the error is a cannot parse error like.
func IsErrCannotParse(err error) bool {
	return errors.Is(err, ErrCannotParseHandle) ||
		errors.Is(err, ErrCannotParseRecordKey) ||
		errors.Is(err, ErrCannotParseURI) ||
		errors.Is(err, ErrCannotParseCID) ||
		errors.Is(err, ErrCannotParseDID) ||
		errors.Is(err, ErrCannotParseNSID) ||
		errors.Is(err, ErrCannotParseTID) ||
		errors.Is(err, ErrCannotParseTime)
}

// ErrDIDNotFound indicates that the [DID] was not found.
//
// Use [errors.As] on [ErrDIDPlcResolve] or on [ErrDIDWebResolve] to get it.
type ErrDIDNotFound struct {
	inner error
}

func (err ErrDIDNotFound) Error() string {
	return "did not found: " + err.inner.Error()
}

func (err ErrDIDNotFound) Unwrap() error {
	return err.inner
}

func (err ErrDIDNotFound) Is(e error) bool {
	var plc ErrDIDPlcResolve
	if errors.As(err, &plc) {
		return plc.StatusCode == http.StatusNotFound
	}
	var web ErrDIDWebResolve
	if errors.As(err, &web) {
		return web.StatusCode == http.StatusNotFound
	}
	return false
}

func handleHTTPError(err error, notFoundErr error) error {
	var dns *net.DNSError
	if errors.As(err, &dns) && dns.IsNotFound {
		return fmt.Errorf("%w: %w", notFoundErr, err)
	}
	return err
}
