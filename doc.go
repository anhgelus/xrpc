// Package xrpc implements a lightweight customizable XRPC [Client].
//
// The [Client] can be used to perform [Client.Query] and [Client.Procedure].
// This returns an [ErrResponse] if the status code is invalid.
// If the API returns a standard error response, you can use [ErrStandardResponse] to retrieve it:
//
//	var err xrpc.ErrResponse
//	e, ok := errors.AsType[xrpc.ErrStandardResponse](err)
//	if ok {
//	  // e is a valid ErrStandardResponse
//	}
//
// Then, you can check standard errors with [ErrStandard]:
//
//	if errors.Is(e, xrpc.ErrRecordNotFound) {
//	  // record not found
//	}
//
// You can use the [Record] interface to define records:
//
//	var CollectionMyRecord = atproto.NewNSIDBuilder("org.example").Name("myRecord").Build()
//
//	type MyRecord struct {
//	  Foo string `json:"bar"`
//	}
//
//	func (r *MyRecord) Collection() *atproto.NSID {
//	  return CollectionMyRecord
//	}
//
// This interface can be used with predefined XRPC functions, like [GetRecord] or [ListRecords].
//
// ## Auth
//
// Some endpoint requires [Auth].
// The [AuthClient] is used to authentificate your requests.
// Currently, only auth using JWT is supported.
//
// You can get a new [AuthClient] using [JWTAuth] with the subpackage [server].
package xrpc
