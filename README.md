# XRPC

Go library implementing a lightweight [XRPC client for the AT Protocol](https://atproto.com/specs/xrpc).

## Why?

The official Bluesky library ([Indigo](https://github.com/bluesky-social/indigo/)) is heavy and use a ton of
dependencies.
In addition to this, their API does not provide an incredible developer experiences.

Use this library if you want a lightweight client.
Use Indigo if you want a feature-complete XRPC implementations.

## Scope

This project wants to provide a lightweight application-agnostic XRPC client.
It reimplements the required foundations of the ATProto.

This library is low-level.
We are just creating an HTTP client with ATProto-specific features.
For example, we do not plan to add a randomized exponential backoff.

If it is possible, we will try to make this library compatible with Indigo.

## Roadmap

- [ ] ATProto foundations in package `atproto`
  - [x] DID
  - [x] NSID
  - [x] TID
  - [x] Record key
  - [ ] Collection
  - [ ] Handle
  - [ ] AT URI
- [ ] [Simple query and procedure](https://atproto.com/specs/xrpc#lexicon-http-endpoints)
  - [ ] Client definition
  - [ ] Lexicon definition
  - [ ] Marshal/Unmarshal
- [ ] [Supports blob](https://atproto.com/specs/xrpc#blob-upload-and-download)
- [ ] [Authentication](https://atproto.com/specs/xrpc#authentication)
- [ ] [Service proxying](https://atproto.com/specs/xrpc#service-proxying)
