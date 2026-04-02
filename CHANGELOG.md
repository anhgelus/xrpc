# Changelog

## v0.2.0 - Blob supports

New features:
- upload blobs
- compat with generated lexicons from Indigo
- admin auth

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
