package atproto

import (
	"encoding/json"
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
	Fragment  string
}

// Is returns true if [NSID] represents the same.
func (n *NSID) Is(nsid *NSID) bool {
	return n.Authority == nsid.Authority && n.Name == nsid.Name && n.Fragment == nsid.Fragment
}

func (n *NSID) String() string {
	var sb strings.Builder
	sb.Grow(len(n.Authority) + len(n.Name) + len(n.Fragment) + 2)
	if n.Authority != "" {
		sb.WriteString(n.Authority)
		sb.WriteRune('.')
	}
	sb.WriteString(n.Name)
	if n.Fragment != "" {
		sb.WriteRune('#')
		sb.WriteString(n.Fragment)
	}
	return sb.String()
}

func (n *NSID) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	nsid, err := ParseNSID(s)
	if err != nil {
		return nil
	}
	*n = *nsid
	return nil
}

func (n *NSID) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *NSID) MarshalMap() (any, error) {
	return n.String(), nil
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
	authority, name, fragment string
}

// NewNSIDBuilder creates a new [NSIDBuilder].
func NewNSIDBuilder(base string) NSIDBuilder {
	if len(base) > NSIDMaxLength {
		panic("invalid base: too long, " + base)
	}
	if !strings.ContainsRune(base, '.') {
		panic("invalid base part: must contains at least one dot, " + base)
	}
	for s := range strings.SplitSeq(base, ".") {
		if !regexpNSIDSegment.MatchString(s) {
			panic("invalid base part: doesn't match, " + s)
		}
	}
	return NSIDBuilder{strings.ToLower(base), "", ""}
}

// SubAuthority adds a sub segment to [NSID.Authority].
// Must not start or end with a dot ('.').
func (b NSIDBuilder) SubAuthority(authority string) NSIDBuilder {
	for s := range strings.SplitSeq(authority, ".") {
		if !regexpNSIDSegment.MatchString(s) {
			panic("invalid authority part: doesn't match, " + s)
		}
	}
	b.authority += "." + strings.ToLower(authority)
	if len(b.authority) > NSIDMaxLength {
		panic("invalid authority: too long, " + b.authority)
	}
	return b
}

// Name sets the [NSID.Name].
// Must not start with a dot ('.').
func (b NSIDBuilder) Name(name string) NSIDBuilder {
	if len(name) > 63 {
		panic("invalid name: too long, " + name)
	}
	if !regexpNSIDName.MatchString(name) {
		panic("invalid name: doesn't match, " + name)
	}
	b.name = name
	return b
}

// Fragment sets the [NSID.Fragment].
// Must not start with a #.
//
// Does not perform any validation, because the documentations is unclear.
func (b NSIDBuilder) Fragment(fragment string) NSIDBuilder {
	b.fragment = fragment
	return b
}

// Build the [NSID].
func (b NSIDBuilder) Build() *NSID {
	if len(b.name) == 0 {
		panic("name not set")
	}
	if len(b.authority) == 0 {
		panic("authority not set")
	}
	return &NSID{b.authority, b.name, b.fragment}
}

func (b NSIDBuilder) String() string {
	return b.authority
}

func (b NSIDBuilder) MarshalJSON() ([]byte, error) {
	panic("cannot marshal NSIDBuilder")
}

func (b NSIDBuilder) UnmarshalJSON(_ []byte) error {
	panic("cannot unmarshal NSIDBuilder")
}
