package atproto

import (
	"errors"
	"regexp"
)

// RecordKey is used to name and reference an individual record within the same collection of an atproto repository.
type RecordKey string

var regexpRecordKey = regexp.MustCompile(`^[a-zA-Z0-9_~.:-]{1,512}$`)

var ErrInvalidRecordKey = errors.New("invalid record key")

// ParseRecordKey in the raw given string.
//
// Returns [ErrInvalidRecordKey] if the [RecordKey] is invalid.
func ParseRecordKey(raw string) (RecordKey, error) {
	if raw == "." || raw == ".." {
		return "", ErrInvalidRecordKey
	}
	if !regexpRecordKey.MatchString(raw) {
		return "", ErrInvalidRecordKey
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
