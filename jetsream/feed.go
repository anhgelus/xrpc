package jetsream

import "tangled.org/anhgelus.world/xrpc/atproto"

// Option used in the [Feed].
type Option struct {
	Collections         []*atproto.NSID
	DIDs                []*atproto.DID
	MaxMessageSizeBytes uint
	Cursor              uint
	RequireHello        bool
}

// Feed is connected to a Jetstream.
type Feed struct{}
