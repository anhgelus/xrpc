package atproto

import "errors"

// ErrCannotParse is a generic error containing a cannot parse error like.
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
