package atproto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const HandleInvalid Handle = "handle.invalid"

const HandleMaxLength = 253

var (
	ErrInvalidHandle       = errors.New("invalid handle")
	ErrHandleNotFound      = errors.New("handle not found")
	ErrCannotResolveHandle = errors.New("cannot resolve handle")
)

// Handle is mutable and human-friendly account username, in the form of a DNS hostname.
//
// See [ParseHandle] to parse an [Handle] from a string.
type Handle string

// ParseHandle in the raw string given.
//
// Returns [ErrInvalidHandle] if the [Handle] is invalid.
func ParseHandle(raw string) (Handle, error) {
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return "", ErrInvalidHandle
	}
	if !regexpNSIDSegment.MatchString(parts[0]) {
		return "", ErrInvalidHandle
	}
	for i, p := range parts[1:] {
		if !regexpNSIDSegment.MatchString(p) {
			return "", ErrInvalidHandle
		}
		if i == len(parts)-2 {
			if p[0] >= '0' && p[0] <= '9' {
				return "", ErrInvalidHandle
			}
			switch p {
			case "local",
				"arpa",
				"invalid",
				"localhost",
				"internal",
				"example",
				"onion",
				"alt":
				return "", ErrInvalidHandle
			}
		}
	}
	return Handle(strings.ToLower(raw)), nil
}

func (h Handle) String() string {
	return string(h)
}

func (h *Handle) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*h, err = ParseHandle(s)
	return err
}

func (h Handle) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h Handle) MarshalMap() (any, error) {
	return h.String(), nil
}

func (h Handle) did(ctx context.Context, client *http.Client, resolver *net.Resolver) (*DID, error) {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ch := make(chan *DID)

	var err error
	go func() {
		res, err := resolver.LookupTXT(ctx, "_atproto."+h.String())
		if err != nil {
			ch <- nil
			return
		}
		did, e := parseDidTxtRec(res)
		if !errors.Is(e, ErrHandleNotFound) {
			err = fmt.Errorf("cannot resolve via DNS records: %w", ErrCannotResolveHandle)
		}
		// avoid blocking goroutine
		select {
		case <-ctx2.Done():
		default:
			ch <- did
		}
	}()

	select {
	case <-ctx2.Done():
	case did := <-ch:
		if did != nil {
			return did, nil
		}
	}

	req, e := http.NewRequest(http.MethodGet, h.String()+"/.well-known/atproto-did", nil)
	fn := func() error {
		if e == nil {
			return nil
		}
		if err != nil {
			return errors.Join(err, fmt.Errorf("cannot resolve via HTTP: %w", e))
		} else {
			return e
		}
	}
	resp, e := client.Do(req.WithContext(ctx))
	if e != nil {
		return nil, fn()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		e = ErrHandleNotFound
		return nil, fn()
	}
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return nil, fn()
	}
	did, e := ParseDID(strings.TrimSpace(string(b)))
	if e == nil {
		return did, nil
	}
	return nil, fmt.Errorf("%w: invalid DID in well-known: %w", ErrCannotResolveHandle, fn())
}

// Directory is used to get [DIDDocument] from [Handle] and [DID].
//
// We highly encourage you to implement your own [Directory] to limit requests with a cache.
// You can use [BaseDirectory] as a base.
//
// Can be used concurrently by multiple goroutines.
//
// See [NewDirectory] to create a new [BaseDirectory].
type Directory interface {
	// ResolveHandle to get the [DIDDocument] associated with.
	//
	// Returns [ErrInvalidHandle] if the [Handle] is invalid (must display [HandleInvalid] in this case).
	// Returns [ErrHandleNotFound] if the [Handle] is not found.
	// Returns [ErrCannotResolveHandle] if the [DID] stored is invalid.
	ResolveHandle(context.Context, Handle) (*DIDDocument, error)
	// ResolveDID returns the [DIDDocument] associated with.
	//
	// Returns [ErrDIDPlcResolve] if the [DIDPlcDirectory] returns an error (only if [DID.Method] is [DIDPlc]).
	// Returns [ErrDIDWebResolve] if the web server returns an error (only if [DID.Method] is [DIDWeb]).
	ResolveDID(context.Context, *DID) (*DIDDocument, error)
}

// BaseDirectory is a simple [Directory].
type BaseDirectory struct {
	client   *http.Client
	resolver *net.Resolver
}

// NewDirectory returns a new [BaseDirectory] with the given [http.Client] (for well-known verification) and
// [net.resolver] (for DNS verification).
func NewDirectory(client *http.Client, resolver *net.Resolver) *BaseDirectory {
	return &BaseDirectory{
		client:   client,
		resolver: resolver,
	}
}

func (d *BaseDirectory) ResolveDID(ctx context.Context, did *DID) (*DIDDocument, error) {
	return did.document(ctx, d.client)
}

func (d *BaseDirectory) ResolveHandle(ctx context.Context, h Handle) (*DIDDocument, error) {
	did, err := h.did(ctx, d.client, d.resolver)
	if err != nil {
		return nil, err
	}
	doc, err := d.ResolveDID(ctx, did)
	if err != nil {
		return nil, err
	}
	res, ok := doc.Handle()
	if !ok || res != h {
		return nil, ErrInvalidHandle
	}
	return doc, nil
}

func parseDidTxtRec(res []string) (*DID, error) {
	for _, s := range res {
		_, did, ok := strings.Cut(s, "did=")
		if ok {
			d, err := ParseDID(did)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid DID in DNS records: %w", ErrCannotResolveHandle, err)
			}
			return d, nil
		}
	}
	return nil, ErrHandleNotFound
}
