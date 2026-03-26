package atproto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// DidMethod represents a valid [Did] method.
type DidMethod string

func (d DidMethod) String() string {
	return string(d)
}

const (
	DidWeb DidMethod = "web"
	DidPlc DidMethod = "plc"
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

const (
	DidIdentifierMaxLength = 2048
	DidPlcDirectory        = "https://plc.directory"
)

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

// Document returns the [DidDocument] of the current [Did].
//
// Returns [ErrDidPlcResolve] if the [DidPlcDirectory] returns an error (only if [Did.Method] is [DidPlc]).
// Returns [ErrDidWebResolve] if the web server returns an error (only if [Did.Method] is [DidWeb]).
func (d *Did) Document(ctx context.Context, client *http.Client) (*DidDocument, error) {
	switch d.Method {
	case DidWeb:
		// https://w3c-ccg.github.io/did-method-web/
		target := strings.ReplaceAll(d.Identifier, ":", "/")
		if !strings.Contains(target, "/") {
			target += "/.well-known"
		}
		target = "https://" + target + "/did.json"
		req, err := http.NewRequest(http.MethodGet, target, nil)

		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			return nil, ErrDidWebResolve{resp.StatusCode, b}
		}
		var d DidDocument
		err = json.Unmarshal(b, &d)
		return &d, err
	case DidPlc:
		req, err := http.NewRequest(http.MethodGet, DidPlcDirectory+"/"+d.String(), nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			var e ErrDidPlcResolve
			err = json.Unmarshal(b, &e)
			if err != nil {
				return nil, fmt.Errorf("cannot unmarshal ErrDidPlcResolve: %w", err)
			}
			return nil, e
		}
		var d DidDocument
		err = json.Unmarshal(b, &d)
		return &d, err
	default:
		return nil, ErrUnsupportedDidMethod
	}
}

// ParseDID in the raw string given.
//
// Returns [ErrInvalidDid] is the DID is not a valid ATProto [Did].
// Returns [ErrUnsupportedDidMethod] if the [DidMethod] is not supported.
//
// See [DidWeb] and [DidPlc] for supported [DidMethod].
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
	case DidWeb, DidPlc:
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

type DidDocument struct {
	ID                 string                  `json:"id"`
	AlsoKnownAs        []string                `json:"alsoKnownAs,omitempty"`
	VerificationMethod []DidVerificationMethod `json:"verificationMethod,omitempty"`
	Service            []DidService            `json:"service,omitempty"`
}

type DidVerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase"`
}

type DidService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// ErrDidPlcResolve is returned by the [DidPlcDirectory].
//
// See [Did.Document].
type ErrDidPlcResolve struct {
	Message string `json:"message,omitempty"`
}

func (e ErrDidPlcResolve) Error() string {
	return e.Message
}

// ErrDidWebResolve is returned by the web server.
//
// See [Did.Document].
type ErrDidWebResolve struct {
	StatusCode int
	Body       []byte
}

func (e ErrDidWebResolve) Error() string {
	if e.Body == nil {
		return fmt.Sprintf("invalid status code while fetching document: %d", e.StatusCode)
	}
	return fmt.Sprintf("%s (status code: %d)", e.Body, e.StatusCode)
}
