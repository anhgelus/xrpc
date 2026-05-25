package jetsream

import (
	"time"

	"tangled.org/anhgelus.world/xrpc"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

// Kind of the event.
type Kind string

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

const (
	CreateOperation = "create"
	UpdateOperation = "update"
	DeleteOperation = "delete"
)

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

type Account struct {
	Active bool         `json:"active"`
	DID    *atproto.DID `json:"did"`
	Seq    int64        `json:"seq"`
	Status *string      `json:"status,omitempty"`
	Time   time.Time    `json:"time"`
}

type Identity struct {
	DID    *atproto.DID    `json:"did"`
	Handle *atproto.Handle `json:"handle,omitempty"`
	Seq    int64           `json:"seq" cborgen:"seq"`
	Time   time.Time       `json:"time"`
}
