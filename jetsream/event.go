package jetsream

import (
	"fmt"
	"time"

	"anhgelus.world/xrpc"
	"anhgelus.world/xrpc/atproto"
)

// Kind of the event.
type Kind string

// List of [Event]'s [Kind].
const (
	CommitKind   = "commit"
	IdentityKind = "identity"
	AccountKind  = "account"
)

// Event for [CommitKind].
type Event struct {
	Kind   Kind         `json:"kind"`
	TimeUs uint64       `json:"time_us"`
	DID    *atproto.DID `json:"did"`
	// Nil if [Kind] is not [CommitKind].
	Commit *Commit `json:"commit,omitempty"`
	// Nil if [Kind] is not [AccountKind].
	Account *Account `json:"account,omitempty"`
	// Nil if [Kind] is not [IdentityKind].
	Identity *Identity `json:"identity,omitempty"`
}

// Operation described by the [Commit].
type Operation string

// List of [Commit]'s [Operation].
const (
	CreateOperation = "create"
	UpdateOperation = "update"
	DeleteOperation = "delete"
)

// Commit is an [Operation] on a [Record].
type Commit struct {
	// Current revision.
	Rev string `json:"rev,omitempty"`
	// Operation described by the [Commit].
	Operation  Operation            `json:"operation,omitempty"`
	Collection *atproto.NSID        `json:"collection,omitempty"`
	RKey       *atproto.RecordKey   `json:"rkey,omitempty"`
	Record     *xrpc.Union          `json:"record,omitempty"`
	CID        *atproto.CIDAsString `json:"cid,omitempty"`
}

// Account is a modification on an atproto account.
type Account struct {
	Active bool         `json:"active"`
	DID    *atproto.DID `json:"did"`
	Seq    int64        `json:"seq"`
	Status *string      `json:"status,omitempty"`
	Time   time.Time    `json:"time"`
}

// Identity is linked with an account, like a modification of an [atproto.Handle].
type Identity struct {
	DID    *atproto.DID    `json:"did"`
	Handle *atproto.Handle `json:"handle,omitempty"`
	Seq    int64           `json:"seq" cborgen:"seq"`
	Time   time.Time       `json:"time"`
}

// ErrInvalidEvent is returned when the [Event] sent by [Feed] is invalid.
//
// Use [InvalidEvent] to create a new instance.
type ErrInvalidEvent struct {
	Event  *Event
	Reason error
}

func (e ErrInvalidEvent) Error() string {
	return fmt.Sprintf("invalid event %v: %v", e.Event, e.Reason)
}

func (e ErrInvalidEvent) Unwrap() error {
	return e.Reason
}

// InvalidEvent creates a new [ErrInvalidEvent] for the given [Event] with the given reason.
func InvalidEvent(e *Event, reason error) ErrInvalidEvent {
	return ErrInvalidEvent{e, reason}
}
