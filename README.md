# XRPC

Go library implementing a lightweight [XRPC client for the AT Protocol](https://atproto.com/specs/xrpc).

Main repository is hosted on [Tangled](https://tangled.org/anhgelus.world/xrpc/), an ATProto forge.

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

## Usage

Get the module with
```bash 
go get -u tangled.org/anhgelus.world/xrpc
```

ATProto primitives are in `atproto` package.

You can create a new simple XRPC client with:
```go
var httpClient *http.Client
client := xrpc.NewClient(httpClient)
```

To create a new request, you can use `client.NewRequest`:
```go
// creating the request
req := client.NewRequest().
    Server("https://..."). // URL of the XRPC server (PDS, relay...)
    Endpoint(atproto.NewNSIDBuilder("org.example").Finish("fooBar")). // XRPC endpoint called
    Params(nil). // optional url.Values
    Build() // can panic if something is wrong
// XRPC query
b, err := client.Query(context.TODO, req)
if err != nil {
    panic(err)
}
// XRPC procedure
body := xrpc.RawBodyRequest{[]byte("Hello world :D"), "text/plain"} // procedure body
b, err := client.Procedure(context.TODO, req, body)
if err != nil {
    panic(err)
}
```

### Using simple records

Another way to interact with an XRPC server is to use the `Record` interface.
It describes an ATProto record (an object based on a lexicon) and must be serialized into JSON and into a map.
```go
type MyRecord struct {
    Hey string `json:"hey"`
}

var myRecordType = atproto.NewNSIDBuilder("org.example").Finish("foo")

func (r *MyRecord) Type() *atproto.NSID {
    return myRecordType
}
```

Then, you can use the higher level API:
```go
// get a record
var did *atproto.DID // did of a user
var rkey atproto.RecordKey // rkey of the record
rec, err := xrpc.GetRecord[*MyRecord](
    context.TODO(), 
    client, 
    "https://...", // URL of the XRPC server (PDS, relay...)
    did,
    rkey,
    ""
)
if err != nil {
    panic(err)
}
// list records
recs, err := xrpc.ListRecords[*MyRecord](
    context.TODO(), 
    client, 
    "https://...", // URL of the XRPC server (PDS, relay...)
    did,
    0, "", false, // not required values
)
if err != nil {
    panic(err)
}
```

### Complexe records

When your record is sent, it is firstly marshaled to a map following the json tags you have defined.
This automatic process may produce invalid results for complexe types.
You can implement `xrpc.MapMarshaler` to specify how it must be marshaled:
```go
func (r *MyRecord) MarshalMap() (any, error) {
    return map[string]any{"foo": r.Hey}, nil
}
```
You can use `xrpc.MarshalToMap` to marshal anything.

The data is sent by the server in JSON.
The library decode it automatically.
You can implement `json.Umarshaler` to specify how it must be unmarshaled:
```go
func (r *MyRecord) UnmarshalJSON(b []byte) error {
    var v struct {
        Hey string `json:"foo"`
    }
    err := json.Unmarshal(b, &v)
    r.Hey = v.Hey
    return err
}
```

### Translating lexicons into Go types

Every ATProto specific type like DID, TID, NSID... must be represented using their definition in the package `atproto`.

An open union must be represented by `*xrpc.Union`.

A custom record can be safely represented by its own type.

```go
type MyComplexeRecord struct {
    URI atproto.RawURI `json:"uri"`
    Record *MyRecord `json:"record"`
    RKey atproto.RecordKey `json:"rkey"`
    Anything *xrpc.Union `json:"anything"`
}
// no need to implement xrpc.MapMarshaler or json.Unmarshaler here :D
```

## Roadmap

- [x] ATProto foundations in package `atproto`
  - [x] DID
  - [x] NSID
  - [x] TID
  - [x] Record key
  - [x] Handle
  - [x] AT URI
- [ ] [Simple query and procedure](https://atproto.com/specs/xrpc#lexicon-http-endpoints)
  - [x] Client definition
  - [x] Record definition
  - [x] Marshal/Unmarshal
- [ ] [Supports blob](https://atproto.com/specs/xrpc#blob-upload-and-download)
- [ ] [Authentication](https://atproto.com/specs/xrpc#authentication)
- [ ] [Service proxying](https://atproto.com/specs/xrpc#service-proxying)
