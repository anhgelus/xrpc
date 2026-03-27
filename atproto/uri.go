package atproto

import (
	"errors"
	"fmt"
	"strings"
)

// Authority is used to identify a user.
// It can be a [DID] or an [Handle].
type Authority interface {
	Handle | *DID
	fmt.Stringer
}

var ErrInvalidURI = errors.New("invalid AT URI")

// RawURI is an [URI] where its [Authority] is not determined.
//
// Use [RawURI.IsDID] and [RawURI.IsHandle] to determine the type.
// Use [RawURI.DID] and [RawURI.Handle] to get the [URI].
//
// See [ParseRawURI] to parse a [RawURI] from a string. 
type RawURI struct {
	raw  string
	kind any
}

// ParseRawURI in the raw given string.
//
// Returns [ErrInvalidURI] if the begining of the [RawURI] is invalid.
// It doesn't verify the syntax: it is when the [RawURI] is converted into an [URI].
func ParseRawURI(raw string) (uri RawURI, err error) {
	uri.raw = raw
	b, raw, ok := strings.Cut(raw, "at://")
	if !ok || b != "" {
		err = ErrInvalidURI
		return
	}
	if strings.HasPrefix(raw, "did:") {
		uri.kind = &DID{}
	} else {
		uri.kind = Handle("")
	}
	return
}

func (r RawURI) IsHandle() bool {
	_, ok := r.kind.(Handle)
	return ok
}

func (r RawURI) IsDID() bool {
	_, ok := r.kind.(*DID)
	return ok
}

func (r RawURI) Handle() (URI[Handle], error) {
	return ParseURI[Handle](r.raw)
}

func (r RawURI) DID() (URI[*DID], error) {
	return ParseURI[*DID](r.raw)
}

// URI is an ATProto URI Scheme (at://).
//
// See [ParseURI] to parse an [URI] from a string.
// See [RawURI] if the [Authority] is unknown.
type URI[T Authority] struct {
	authority  T
	collection *NSID
	recordKey  *RecordKey
}

// ParseURI in the raw given string.
//
// Returns [ErrInvalidURI] if the [URI] is invalid.
// Returns [ErrCannotParseURIAs] if the [URI] is not compatible with the provided kind.
func ParseURI[T Authority](raw string) (uri URI[T], err error) {
	// parsing authority
	b, raw, ok := strings.Cut(raw, "at://")
	if !ok || b != "" {
		err = ErrInvalidURI
		return
	}
	var t any = uri.authority
	authority, next, ok := strings.Cut(raw, "/")
	if ok && next == "" {
		err = ErrInvalidURI
		return
	}
	switch t.(type) {
	case *DID:
		t, err = ParseDID(authority)
	case Handle:
		t, err = ParseHandle(authority)
	default:
		panic("unsupported authority")
	}
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrCannotParseURIAs{uri}, err)
		return
	}
	uri.authority, ok = t.(T)
	if !ok {
		err = ErrCannotParseURIAs{uri}
		return
	}
	if next == "" {
		return
	}
	// parsing collection
	parts := strings.Split(next, "/")
	if len(parts) > 2 {
		err = ErrInvalidURI
		return
	}
	uri.collection, err = ParseNSID(parts[0])
	if len(parts) == 1 || err != nil {
		return
	}
	// parsing rkey
	var rkey RecordKey
	rkey, err = ParseRecordKey(parts[1])
	if err != nil {
		return
	}
	uri.recordKey = &rkey
	return
}

func (u URI[T]) String() string {
	var sb strings.Builder
	sb.WriteString("at://")
	sb.WriteString(u.Authority().String())
	if u.collection != nil {
		sb.WriteRune('/')
		sb.WriteString(u.Collection().String())
		if u.recordKey != nil {
			sb.WriteRune('/')
			sb.WriteString(u.RecordKey().String())
		}
	}
	return sb.String()
}

func (u URI[T]) Authority() T {
	return u.authority
}

func (u URI[T]) Collection() *NSID {
	return u.collection
}

func (u URI[T]) SetCollection(collection *NSID) URI[T] {
	u.collection = collection
	return u
}

func (u URI[T]) RecordKey() *RecordKey {
	return u.recordKey
}

func (u URI[T]) SetRecordKey(rkey RecordKey) URI[T] {
	u.recordKey = &rkey
	return u
}

type ErrCannotParseURIAs struct {
	target any
}

func (err ErrCannotParseURIAs) Error() string {
	return fmt.Sprintf("cannot parse URI as %T", err.target)
}
