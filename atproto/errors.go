package atproto

import (
	"errors"
	"net"
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
	return "DID not found: " + err.inner.Error()
}

func (err ErrDIDNotFound) Unwrap() error {
	return err.inner
}

func handleHTTPError(err error, notFoundErr error) error {
	var dns *net.DNSError
	if errors.As(err, &dns) && dns.IsNotFound {
		return notFoundErr
	}
	return err
}
