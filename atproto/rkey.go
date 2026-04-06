package atproto

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// RecordKey is used to name and reference an individual record within the same collection of an atproto repository.
type RecordKey string

var regexpRecordKey = regexp.MustCompile(`^[a-zA-Z0-9_~.:-]{1,512}$`)

// Errors returned while parsing a [RecordKey].
var (
	ErrNotRecordKey = errors.New("not a record key")
	// ErrCannotParseRecordKey is returned by [ParseRecordKey] if an error occurs.
	ErrCannotParseRecordKey = errors.New("cannot parse RecordKey")
)

// ParseRecordKey in the raw given string.
//
// Returns [ErrNotRecordKey] if the [RecordKey] is invalid.
func ParseRecordKey(raw string) (RecordKey, error) {
	if raw == "." || raw == ".." {
		return "", fmt.Errorf("%w: %w", ErrCannotParseRecordKey, ErrNotRecordKey)
	}
	if !regexpRecordKey.MatchString(raw) {
		return "", fmt.Errorf("%w: %w", ErrCannotParseRecordKey, ErrNotRecordKey)
	}
	return RecordKey(raw), nil
}

func (r RecordKey) String() string {
	return string(r)
}

func (r RecordKey) TID() (TID, error) {
	return ParseTID(r.String())
}

func (r RecordKey) NSID() (*NSID, error) {
	return ParseNSID(r.String())
}

func (r RecordKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r RecordKey) MarshalMap() (any, error) {
	return r.String(), nil
}

func (r *RecordKey) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r, err = ParseRecordKey(s)
	return err
}
