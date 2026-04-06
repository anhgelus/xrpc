# Changelog

## v0.4.0

Breaking change:
- rename `atproto.ErrInvalid...` into `atproto.ErrNot...`

New features:
- simplify errors handling

## v0.3.0

Breaking change:
- remove `atproto.Handle.DID()`, `atproto.Handle.PDS()` and `atproto.DID.PDS()`: use `atproto.Directory` and `atproto.DIDDocument` instead

New features:
- client to communicate with a relay instead of a PDS

Fix:
- invalid error set in lookupHandle

## v0.2.0 - Blob supports

New features:
- upload blobs
- compat with generated lexicons from Indigo
- admin auth
- create invite code
- specify user agent in client

Fix:
- get return nil in value

## v0.1.0 - First release

First release of XRPC!

Includes:
- ATProto primitives (lexicon-agnostic) in package `atproto`
- XRPC client definition in package `xrpc` (root)
- Record definition in package `xrpc` (root)
- Common functions `com.atproto.repo.*` in package `xrpc` (root)
- JWT Auth functions in package `server`
