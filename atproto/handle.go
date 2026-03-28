package atproto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
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

func (h Handle) URI() URI[Handle] {
	return URI[Handle]{authority: h}
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
	return []byte(h.String()), nil
}

func (h Handle) MarshalMap() (any, error) {
	return h.String(), nil
}

// Directory is used to get [DID] from [Handle].
//
// Can be used concurrently by multiple goroutines.
//
// Use [Directory.ResolveHandle] to retrieve the [DIDDocument] associated with an [Handle].
//
// See [NewDirectory] to create a new [Directory].
type Directory struct {
	mu        sync.RWMutex
	client    *http.Client
	resolver  *net.Resolver
	cache     map[Handle]*DIDDocument
	cachedFor time.Duration
}

// NewDirectory returns a new [Directory] with the given [http.Client] (for well-known verification), [net.resolver]
// (for DNS verification) and [time.Duration] (for the time cached in a map).
func NewDirectory(client *http.Client, resolver *net.Resolver, cachedFor time.Duration) *Directory {
	return &Directory{
		client:    client,
		resolver:  resolver,
		cache:     make(map[Handle]*DIDDocument),
		cachedFor: cachedFor,
	}
}

// ResolveHandle to get the [DIDDocument] associated with.
//
// Returns [ErrInvalidHandle] if the [Handle] is invalid.
// Returns [ErrHandleNotFound] if the [Handle] is not found.
// Returns [ErrCannotResolveHandle] if the [DID] stored is invalid.
func (d *Directory) ResolveHandle(ctx context.Context, h Handle) (*DIDDocument, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if doc, ok := d.cache[h]; ok {
		return doc, nil
	}
	did, err := d.lookupHandle(ctx, h)
	if err != nil {
		return nil, err
	}
	doc, err := did.Document(ctx, d.client)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(doc.AlsoKnownAs, h.URI().String()) {
		return nil, ErrInvalidHandle
	}
	d.mu.RUnlock()
	d.mu.Lock()
	d.cache[h] = doc
	go func(h Handle) {
		time.Sleep(d.cachedFor * time.Minute)
		delete(d.cache, h)
	}(h)
	d.mu.Unlock()
	d.mu.RLock()
	return doc, nil
}

func (d *Directory) lookupHandle(ctx context.Context, h Handle) (*DID, error) {
	res, err := d.resolver.LookupTXT(ctx, "_atproto."+h.String())
	if err == nil {
		did, e := parseDidTxtRec(res)
		if e == nil {
			return did, nil
		}
		if !errors.Is(e, ErrHandleNotFound) {
			err = fmt.Errorf("cannot resolve via DNS records: %w", ErrCannotResolveHandle)
		}
	}
	req, e := http.NewRequest(http.MethodGet, h.String()+"/.well-known/atproto-did", nil)
	defer func() {
		if e == nil {
			return
		}
		if err != nil {
			err = errors.Join(err, fmt.Errorf("cannot resolve via HTTP: %w", e))
		} else {
			err = e
		}
	}()
	resp, e := d.client.Do(req.WithContext(ctx))
	if e != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		e = ErrHandleNotFound
		return nil, err
	}
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return nil, err
	}
	did, e := ParseDID(strings.TrimSpace(string(b)))
	if e == nil {
		return did, nil
	}
	return nil, fmt.Errorf("%w: invalid DID in well-known: %w", ErrCannotResolveHandle, err)
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
