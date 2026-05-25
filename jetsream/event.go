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

// EventBase is the data shared in every event.
type EventBase struct {
	Kind   Kind         `json:"kind"`
	TimeUs uint64       `json:"time_us"`
	DID    *atproto.DID `json:"did"`
}

// EventCommit is an event for [CommitKind].
type EventCommit struct {
	EventBase
	Commit *Commit `json:"commit"`
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

// EventAccount is an event for [AccountKind].
type EventAccount struct {
	EventBase
	Account *Account `json:"account"`
}

type Account struct {
	Active bool         `json:"active"`
	DID    *atproto.DID `json:"did"`
	Seq    int64        `json:"seq"`
	Status *string      `json:"status,omitempty"`
	Time   time.Time    `json:"time"`
}

// EventIdentity is an event for [IdentityKind].
type EventIdentity struct {
	EventBase
	Identity *Identity `json:"identity"`
}

type Identity struct {
	DID    *atproto.DID    `json:"did"`
	Handle *atproto.Handle `json:"handle,omitempty"`
	Seq    int64           `json:"seq" cborgen:"seq"`
	Time   time.Time       `json:"time"`
}
