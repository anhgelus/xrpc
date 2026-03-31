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
// ## Record
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
// When your [Record] is sent, it is firstly marshaled into a map with [MarshalToMap] and then marshaled into JSON with
// [json.Marshal].
// If your [Record] requires a custom logic to be marshaled, it must implement [MapMarshaler] and returns a
// map[string]any.
// If you define custom objects, it can return anything.
//
// When your [Record] is received, it is unmarshaled with [json.Unmarshal].
// If it requires a custom logic, you must implement [json.Unmarshaler].
//
// If you want to marshal your [Record] into a JSON, you must use [Marshal].
//
// ## Auth
//
// Some endpoint requires [Auth].
// The [AuthClient] is used to authentificate your requests.
// Currently, only auth using JWT is supported.
//
// You can get a new [AuthClient] using [JWTAuth] with the subpackage [server].
package xrpc
