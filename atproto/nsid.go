package atproto

import (
	"errors"
	"regexp"
	"strings"
)

// NSID represents a namespaced identifier.
//
// See [ParseNSID] to parse a [NSID] from a string.
// See [NSIDBuilder] to create dynamically a [NSID].
type NSID struct {
	Authority string
	Name      string
}

func (n NSID) String() string {
	return n.Authority + "." + n.Name
}

var (
	ErrInvalidNSID = errors.New("invalid NSID")
)

var (
	regexpNSIDSegment = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	regexpNSIDName    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{0,62}$`)
)

const NSIDMaxLength = 253

// ParseNSID in the raw string given.
//
// Returns [ErrInvalidNSID] if the [NSID] is invalid.
//
// See [NSIDBuilder] to create dynamically a [NSID].
func ParseNSID(raw string) (*NSID, error) {
	parts := strings.Split(raw, ".")
	if len(parts) < 3 {
		return nil, ErrInvalidNSID
	}
	tld := parts[0]
	if !regexpNSIDSegment.MatchString(tld) {
		return nil, ErrInvalidNSID
	}
	if tld[0] >= '0' && tld[0] <= '9' {
		return nil, ErrInvalidNSID
	}
	var nsid NSID
	var sb strings.Builder
	sb.WriteString(tld)
	for i, p := range parts[1:] {
		if i == len(parts)-2 {
			if !regexpNSIDName.MatchString(p) {
				return nil, ErrInvalidNSID
			}
			nsid.Name = p
		} else {
			if !regexpNSIDSegment.MatchString(p) {
				return nil, ErrInvalidNSID
			}
			sb.WriteRune('.')
			sb.WriteString(p)
		}
	}
	nsid.Authority = sb.String()
	if len(nsid.Authority) > NSIDMaxLength {
		return nil, ErrInvalidNSID
	}
	nsid.Authority = strings.ToLower(nsid.Authority)
	return &nsid, nil
}

// NSIDBuilder helps creating an [NSID].
// It panics instead of returning an error if an invalid argument is given.
type NSIDBuilder struct {
	authority string
}

// NewNSIDBuilder creates a new [NSIDBuilder].
func NewNSIDBuilder(base string) NSIDBuilder {
	if len(base) > NSIDMaxLength {
		panic("invalid base: " + base)
	}
	for s := range strings.SplitSeq(base, ".") {
		if !regexpNSIDSegment.MatchString(s) {
			panic("invalid base part: " + s)
		}
	}
	return NSIDBuilder{strings.ToLower(base)}
}

// Add a segment to [NSID.Authority].
// Must not start or end with a dot ('.').
func (b NSIDBuilder) Add(authority string) NSIDBuilder {
	for s := range strings.SplitSeq(authority, ".") {
		if !regexpNSIDSegment.MatchString(s) {
			panic("invalid authority part: " + s)
		}
	}
	b.authority += "." + strings.ToLower(authority)
	if len(b.authority) > NSIDMaxLength {
		panic("invalid authority: " + b.authority)
	}
	return b
}

// Finish constructing an [NSID] by setting the [NSID.Name].
func (b NSIDBuilder) Finish(name string) *NSID {
	if len(name) > 63 {
		panic("invalid name: " + name)
	}
	if len(b.authority) == 0 {
		panic("authority not set")
	}
	if !regexpNSIDName.MatchString(name) {
		panic("invalid name: " + name)
	}
	return &NSID{b.authority, name}
}

func (b NSIDBuilder) String() string {
	return b.authority
}
