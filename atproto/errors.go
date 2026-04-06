package atproto

import (
	"errors"
	"net/http"
)

// ErrCannotParse is a generic error containing a cannot parse error like.
//
// See [AsCannotParse] to converts an error into [ErrCannotParse].
// The standard [errors.As] does not work: this is mainly a meta error.
type ErrCannotParse struct {
	inner error
}

func (err ErrCannotParse) Error() string {
	return err.inner.Error()
}

func (err ErrCannotParse) Unwrap() error {
	return err.inner
}

func (err ErrCannotParse) Is(e error) bool {
	if _, ok := e.(ErrCannotParse); ok {
		return true
	}
	return errors.Is(e, ErrCannotParseHandle) ||
		errors.Is(e, ErrCannotParseRecordKey) ||
		errors.Is(e, ErrCannotParseURI) ||
		errors.Is(e, ErrCannotParseCID) ||
		errors.Is(e, ErrCannotParseDID) ||
		errors.Is(e, ErrCannotParseNSID) ||
		errors.Is(e, ErrCannotParseTID) ||
		errors.Is(e, ErrCannotParseTime)
}

// AsCannotParse converts an error into [ErrCannotParse].
//
// Panics if [ErrCannotParse.Is] returns false.
func AsCannotParse(err error) ErrCannotParse {
	if !errors.Is(err, ErrCannotParse{}) {
		panic("cannot convert " + err.Error() + " to parse error")
	}
	return ErrCannotParse{err}
}

// ErrDIDNotFound indicates that the [DID] was not found.
//
// Use [errors.As] on [ErrDIDPlcResolve] or on [ErrDIDWebResolve] to get it.
type ErrDIDNotFound struct {
	inner error
}

func (err ErrDIDNotFound) Error() string {
	return err.inner.Error()
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
