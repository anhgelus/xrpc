package atproto

import (
	"context"
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

// Identifier is used to identify a user.
// It can be a [DID] or an [Handle].
type Identifier interface {
	Handle | DID
	URI() string
}

// Handle is mutable and human-friendly account username, in the form of a DNS hostname.
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

func (h Handle) URI() string {
	return "at://" + h.String()
}

type Directory struct {
	mu        sync.RWMutex
	client    *http.Client
	resolver  *net.Resolver
	cache     map[Handle]*DIDDocument
	cachedFor time.Duration
}

func NewDirectory(client *http.Client, resolver *net.Resolver, cachedFor time.Duration) *Directory {
	return &Directory{
		client:    client,
		resolver:  resolver,
		cache:     make(map[Handle]*DIDDocument),
		cachedFor: cachedFor,
	}
}

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
	if !slices.Contains(doc.AlsoKnownAs, h.URI()) {
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
	return nil, err
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
