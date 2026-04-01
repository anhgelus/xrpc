// Package atproto implements AT Protocol primitives.
//
// It aims to include every lexicon-agnostic definition to work everywhere, including outside of the standard
// ATmosphere.
// Each type can be safely marshaled into JSON or into a map using the [xrpc] package.
// Each type can be safely unmarshaled from json too.
//
// When working with [Handle] and [DID], you will probably need a [Directory].
// It is an interface used to retrieve [DIDDocument].
// We strongly encourage you to implement your own [Directory] to include caching.
//
// When working with [URI], we have decided to normalize every [Handle] used as authority into their [DID].
// For example, at://anhgelus.world become at://did:plc:vtqucb4iga7b5wzza3zbz4so.
// You can use a [RawURI] to avoid this (like in unmarshaled struct), but it does not perform any checks.
//
// To create [TID], we strongly encourage you to use [TIDGenerator] to generates [TID] that always increase.
package atproto
