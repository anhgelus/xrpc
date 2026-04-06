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

// DIDMethod represents a valid [DID] method.
type DIDMethod string

func (d DIDMethod) String() string {
	return string(d)
}

// Supported [DIDMethod].
const (
	DIDWeb DIDMethod = "web"
	DIDPlc DIDMethod = "plc"
)

var (
	regexpDidMethod     = regexp.MustCompile(`^[a-z]+$`)
	regexpDidIdentifier = regexp.MustCompile(`^[a-zA-Z0-9._:%-]*[a-zA-Z0-9._-]$`)
)

func asDIDMethod(raw string) (DIDMethod, bool) {
	if !regexpDidMethod.MatchString(raw) {
		return "", false
	}
	return DIDMethod(raw), true
}

const (
	DIDIdentifierMaxLength = 2048
	DIDPlcDirectory        = "https://plc.directory"
)

// Errors returned while parsing a [DID].
var (
	ErrNotDID               = errors.New("not a DID")
	ErrUnsupportedDIDMethod = errors.New("unsupported DID method")
	// ErrCannotParseDID is returned by [ParseDID] if an error occurs.
	ErrCannotParseDID = errors.New("cannot parse DID")
)

// DID represents a W3C DID in the context of the ATProto.
//
// See [ParseDID] to parse a [DID] from a string.
type DID struct {
	Method     DIDMethod
	Identifier string
}

func (d *DID) String() string {
	return "did:" + string(d.Method) + ":" + d.Identifier
}

func (d *DID) URI() URI {
	return URI{authority: d}
}

func (d *DID) document(ctx context.Context, client *http.Client) (*DIDDocument, error) {
	switch d.Method {
	case DIDWeb:
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
			err = handleHTTPError(err, ErrCannotParseDID)
			return nil, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			return nil, ErrDIDWebResolve{resp.StatusCode, b}
		}
		var d DIDDocument
		err = json.Unmarshal(b, &d)
		return &d, err
	case DIDPlc:
		req, err := http.NewRequest(http.MethodGet, DIDPlcDirectory+"/"+d.String(), nil)
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
			var e ErrDIDPlcResolve
			err = json.Unmarshal(b, &e)
			if err != nil {
				return nil, fmt.Errorf("cannot unmarshal ErrDidPlcResolve: %w", err)
			}
			e.StatusCode = resp.StatusCode
			return nil, e
		} else if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("unknown PLC error status code: %s (status code: %d)", b, resp.StatusCode)
		}
		var d DIDDocument
		err = json.Unmarshal(b, &d)
		return &d, err
	default:
		return nil, ErrUnsupportedDIDMethod
	}
}

func (d *DID) MarshalMap() (any, error) {
	return d.String(), nil
}

func (d *DID) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *DID) UnmarshalJSON(b []byte) error {
	var raw string
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	did, err := ParseDID(raw)
	if err != nil {
		return err
	}
	*d = *did
	return nil
}

// ParseDID in the raw string given.
//
// Returns [ErrNotDID] is the DID is not a valid ATProto [DID].
// Returns [ErrUnsupportedDIDMethod] if the [DIDMethod] is not supported.
//
// See [DIDWeb] and [DIDPlc] for supported [DIDMethod].
func ParseDID(raw string) (*DID, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 3 || parts[0] != "did" {
		return nil, fmt.Errorf("%w: %w", ErrCannotParseDID, ErrNotDID)
	}
	var d DID
	var ok bool
	d.Method, ok = asDIDMethod(parts[1])
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrCannotParseDID, ErrNotDID)
	}
	switch d.Method {
	case DIDWeb, DIDPlc:
	default:
		return nil, fmt.Errorf("%w: %w", ErrCannotParseDID, ErrUnsupportedDIDMethod)
	}
	d.Identifier = parts[2]
	if !regexpDidIdentifier.MatchString(d.Identifier) {
		return nil, fmt.Errorf("%w: %w", ErrCannotParseDID, ErrNotDID)
	}
	if len([]rune(d.Identifier)) > DIDIdentifierMaxLength {
		return nil, fmt.Errorf("%w: %w", ErrCannotParseDID, ErrNotDID)
	}
	return &d, nil
}

// DIDDocument stores information about a [DID].
type DIDDocument struct {
	DID                *DID                    `json:"id"`
	AlsoKnownAs        []string                `json:"alsoKnownAs,omitempty"`
	VerificationMethod []DIDVerificationMethod `json:"verificationMethod,omitempty"`
	Service            []DIDService            `json:"service,omitempty"`
}

func (d *DIDDocument) PDS() (string, bool) {
	for _, s := range d.Service {
		if s.ID == "#atproto_pds" && s.Type == "AtprotoPersonalDataServer" {
			return s.ServiceEndpoint, true
		}
	}
	return "", false
}

func (d *DIDDocument) Handle() (Handle, bool) {
	for _, as := range d.AlsoKnownAs {
		b, n, ok := strings.Cut(as, "at://")
		if ok && b == "" {
			h, err := ParseHandle(n)
			return h, err == nil
		}
	}
	return "", false
}

type DIDVerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase"`
}

type DIDService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// ErrDIDPlcResolve is returned by the [DIDPlcDirectory].
//
// See [DID.document].
type ErrDIDPlcResolve struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message,omitempty"`
}

func (e ErrDIDPlcResolve) Error() string {
	return fmt.Sprintf("%s (status code: %d)", e.Message, e.StatusCode)
}

func (e ErrDIDPlcResolve) Is(err error) bool {
	switch err.(type) {
	case ErrDIDPlcResolve:
		return true
	case ErrDIDNotFound:
		return e.StatusCode == http.StatusNotFound
	default:
		return false
	}
}

func (e ErrDIDPlcResolve) As(target any) bool {
	v, ok := target.(*ErrDIDNotFound)
	if !ok || e.StatusCode == http.StatusNotFound {
		return false
	}
	*v = ErrDIDNotFound{e}
	return true
}

// ErrDIDWebResolve is an error returned by the [DIDPlcDirectory].
//
// See [Directory.ResolveDID] and [Directory.ResolveHandle].
type ErrDIDWebResolve struct {
	StatusCode int
	Body       []byte
}

func (e ErrDIDWebResolve) Error() string {
	if e.Body == nil {
		return fmt.Sprintf("invalid status code while fetching document: %d", e.StatusCode)
	}
	return fmt.Sprintf("%s (status code: %d)", e.Body, e.StatusCode)
}

func (e ErrDIDWebResolve) Is(err error) bool {
	switch err.(type) {
	case ErrDIDWebResolve:
		return true
	case ErrDIDNotFound:
		return e.StatusCode == http.StatusNotFound
	default:
		return false
	}
}

func (e ErrDIDWebResolve) As(target any) bool {
	v, ok := target.(*ErrDIDNotFound)
	if !ok || e.StatusCode == http.StatusNotFound {
		return false
	}
	*v = ErrDIDNotFound{e}
	return true
}
