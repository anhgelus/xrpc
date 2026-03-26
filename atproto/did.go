package atproto

import (
	"errors"
	"regexp"
	"strings"
)

type DidMethod string

func (d DidMethod) String() string {
	return string(d)
}

const (
	DidWeb DidMethod = "web"
	DidPLC DidMethod = "plc"
)

var (
	regexpDidMethod     = regexp.MustCompile(`[a-z]+`)
	regexpDidIdentifier = regexp.MustCompile(`[a-zA-Z0-9._:%-]*[a-zA-Z0-9._-]`)
)

func asDidMethod(raw string) (DidMethod, bool) {
	if !regexpDidMethod.MatchString(raw) {
		return "", false
	}
	return DidMethod(raw), true
}

const DidIdentifierMaxLength = 2048

var (
	ErrInvalidDid           = errors.New("invalid DID")
	ErrUnsupportedDidMethod = errors.New("unsupported DID method")
)

// Did represents a DID in the context of the ATProto.
//
// See [ParseDID] to parse a [Did].
type Did struct {
	Method     DidMethod
	Identifier string
}

func (d *Did) String() string {
	return "did:" + string(d.Method) + ":" + d.Identifier
}

// ParseDID in the raw string given.
//
// Returns [ErrInvalidDid] is the DID is not a valid ATProto [Did].
// Returns [ErrUnsupportedDidMethod] if the [DidMethod] is not supported.
//
// See [DidWeb] and [DidPLC] for supported [DidMethod].
func ParseDID(raw string) (*Did, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 3 {
		return nil, ErrInvalidDid
	}
	if parts[0] != "did" {
		return nil, ErrInvalidDid
	}
	var d Did
	var ok bool
	d.Method, ok = asDidMethod(parts[1])
	if !ok {
		return nil, ErrInvalidDid
	}
	switch d.Method {
	case DidWeb, DidPLC:
	default:
		return nil, ErrUnsupportedDidMethod
	}
	d.Identifier = parts[2]
	if !regexpDidIdentifier.MatchString(d.Identifier) {
		return nil, ErrInvalidDid
	}
	if len([]rune(d.Identifier)) > DidIdentifierMaxLength {
		return nil, ErrInvalidDid
	}
	return &d, nil
}
