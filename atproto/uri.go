package atproto

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

var ErrCannotFindPDS = errors.New("cannot find PDS")

var ErrInvalidURI = errors.New("invalid AT URI")

// RawURI is an [URI] where its [Authority] is not determined.
//
// Use [RawURI.IsDID] and [RawURI.IsHandle] to determine the type.
// Use [RawURI.URI] and [RawURI.Handle] to get the [URI].
//
// See [ParseRawURI] to parse a [RawURI] from a string.
type RawURI struct {
	raw string
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
	return
}

func (r RawURI) String() string {
	return r.raw
}

func (r RawURI) URI(ctx context.Context, dir *Directory) (URI, error) {
	return ParseURI(ctx, dir, r.raw)
}

func (r RawURI) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r RawURI) MarshalMap() (any, error) {
	return r.String(), nil
}

func (r *RawURI) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r, err = ParseRawURI(s)
	return err
}

// URI is an ATProto URI Scheme (at://).
//
// See [ParseURI] to parse an [URI] from a string.
// See [RawURI] if the [Authority] is unknown, like in JSON.
// See [NewURI] to create a new [URI].
type URI struct {
	authority  *DID
	collection *NSID
	recordKey  *RecordKey
}

// NewURI creates a new [URI].
func NewURI(authority *DID, collection *NSID, rkey RecordKey) URI {
	return URI{authority, collection, &rkey}
}

// ParseURI in the raw given string.
//
// Returns [ErrInvalidURI] if the [URI] is invalid.
// Returns [ErrCannotParseURIAs] if the [URI] is not compatible with the provided kind.
func ParseURI(ctx context.Context, dir *Directory, raw string) (uri URI, err error) {
	// parsing authority
	b, raw, ok := strings.Cut(raw, "at://")
	if !ok || b != "" {
		err = ErrInvalidURI
		return
	}
	authority, next, ok := strings.Cut(raw, "/")
	if ok && next == "" {
		err = ErrInvalidURI
		return
	}
	defer func() {
		if err != nil {
			return
		}
		uri.authority, err = ParseDID(authority)
		if !strings.HasPrefix(authority, "did:") && err != nil {
			var h Handle
			h, err = ParseHandle(authority)
			if err != nil {
				return
			}
			var doc *DIDDocument
			doc, err = dir.ResolveHandle(ctx, h)
			if err != nil {
				return
			}
			uri.authority = doc.DID
		}
	}()
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

func (u URI) String() string {
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

func (u URI) Authority() *DID {
	return u.authority
}

func (u URI) Collection() *NSID {
	return u.collection
}

func (u URI) SetCollection(collection *NSID) URI {
	u.collection = collection
	return u
}

func (u URI) RecordKey() *RecordKey {
	return u.recordKey
}

func (u URI) SetRecordKey(rkey RecordKey) URI {
	u.recordKey = &rkey
	return u
}
