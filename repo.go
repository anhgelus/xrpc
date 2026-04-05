package xrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"tangled.org/anhgelus.world/xrpc/atproto"
)

// RecordStored represents a [Record] containg values about how it is stored.
type RecordStored[T Record] struct {
	Value T                    `json:"value"`
	URI   atproto.RawURI       `json:"uri"`
	CID   *atproto.CIDAsString `json:"cid"`
}

var (
	ErrRecordNotFound = ErrStandard("RecordNotFound")
)

// GetRecord returns a single [Record] from a repository.
//
// If cid is omitted, it will return the latest version of the [Record].
func GetRecord[T Record](
	ctx context.Context,
	client Client,
	did *atproto.DID,
	rkey atproto.RecordKey,
	cid *atproto.CID,
) (RecordStored[T], error) {
	var v RecordStored[T]
	b, err := rawGetRecord(ctx, client, did, v.Value.Collection(), rkey, cid)
	if err != nil {
		return v, err
	}
	return v, json.Unmarshal(b, &v)
}

func rawGetRecord(
	ctx context.Context,
	client Client,
	did *atproto.DID,
	col *atproto.NSID,
	rkey atproto.RecordKey,
	cid *atproto.CID,
) ([]byte, error) {
	params := make(url.Values)
	params.Add("repo", did.String())
	params.Add("collection", col.String())
	params.Add("rkey", rkey.String())
	if cid != nil {
		params.Add("cid", cid.String())
	}
	req, err := client.NewRequest().
		PDS(ctx, client.Directory(), did)
	if err != nil {
		return nil, err
	}
	req = req.Endpoint(collection.Name("getRecord").Build()).
		Params(params)
	return client.Query(ctx, req)
}

type listOut[T Record] struct {
	Cursor  string            `json:"cursor,omitempty"`
	Records []RecordStored[T] `json:"records"`
}

// ListRecords in a repository.
//
// limit is optional (default: 50, max: 100), set to -1 to use default.
// cursor is optional.
func ListRecords[T Record](
	ctx context.Context,
	client Client,
	did *atproto.DID,
	limit uint8,
	cursor string,
	reverse bool,
) ([]RecordStored[T], string, error) {
	var v T
	params := make(url.Values)
	params.Add("repo", did.String())
	params.Add("collection", v.Collection().String())
	if limit == 0 {
		limit = 50
	}
	params.Add("limit", strconv.Itoa(int(min(limit, 100))))
	if cursor != "" {
		params.Add("cursor", cursor)
	}
	params.Add("reverse", fmt.Sprintf("%t", reverse))
	req, err := client.NewRequest().
		PDS(ctx, client.Directory(), did)
	if err != nil {
		return nil, "", err
	}
	req = req.Endpoint(collection.Name("listRecords").Build()).
		Params(params)
	b, err := client.Query(ctx, req)
	if err != nil {
		return nil, "", err
	}
	var out listOut[T]
	err = json.Unmarshal(b, &out)
	return out.Records, out.Cursor, err
}

// RepoDescription is returned by [DescribeRepo].
type RepoDescription struct {
	Handle atproto.Handle       `json:"handle"`
	DID    *atproto.DID         `json:"did"`
	DIDDoc *atproto.DIDDocument `json:"didDoc"`
	// Collections for which this repo contains at least one [Record].
	Collections     []*atproto.NSID `json:"collections"`
	HandleIsCorrect bool            `json:"handleIsCorrect"`
}

// DescribeRepo returns information about an account and repository.
func DescribeRepo(ctx context.Context, client Client, did *atproto.DID) (*RepoDescription, error) {
	params := make(url.Values)
	params.Add("repo", did.String())
	req, err := client.NewRequest().
		PDS(ctx, client.Directory(), did)
	if err != nil {
		return nil, err
	}
	req = req.Endpoint(collection.Name("describeRepo").Build()).
		Params(params)
	b, err := client.Query(ctx, req)
	if err != nil {
		return nil, err
	}
	var v RepoDescription
	return &v, json.Unmarshal(b, &v)
}

type sendRecordRequest struct {
	Rec        Record               `json:"record,omitempty"`
	Repo       *atproto.DID         `json:"repo"`
	Collection *atproto.NSID        `json:"collection"`
	RKey       atproto.RecordKey    `json:"rkey,omitempty"`
	Validate   *bool                `json:"validate,omitempty"`
	SwapCommit *atproto.CIDAsString `json:"swapCommit,omitempty"`
	SwapRecord *atproto.CIDAsString `json:"swapRecord,omitempty"`
}

// ErrInvalidSwap indicates that 'swap' didn't match current repo commit.
var ErrInvalidSwap = ErrStandard("InvalidSwap")

type RecordValidationStatus string

const (
	RecordValidationValid   RecordValidationStatus = "valid"
	RecordValidationUnknown RecordValidationStatus = "unknown"
)

// SendRecordResult is returned by [CreateRecord] and [PutRecord].
type SendRecordResult struct {
	// URI of the [Record].
	URI *atproto.RawURI `json:"uri"`
	// CID of the [Record].
	CID *atproto.CIDAsString `json:"cid"`
	// Commit infromation about the [Record].
	Commit *CommitMeta `json:"commitMeta,omitempty"`
	// ValidationStatus indicates how the [Record] was validated.
	ValidationStatus RecordValidationStatus `json:"validationStatus,omitempty"`
}

// CreateRecord for the authentificated [Client].
//
// Rkey used to store the [Record] (if unset, autopopulated by default).
// Validate indicates whether the server should validate the [Record] against the Lexicon schema (if unset, validate
// only for known lexicons).
// Swap is used to compare and swap with the previous commit (not required, disabled by default).
func CreateRecord(
	ctx context.Context,
	client Client,
	rec Record,
	rkey atproto.RecordKey,
	validate *bool,
	swap *atproto.CID,
) (*SendRecordResult, error) {
	req := client.NewRequest().Endpoint(collection.Name("createRecord").Build())
	return sendRecord(ctx, client, req, sendRecordRequest{
		Rec:        rec,
		Repo:       req.GetAuth().DID(),
		Collection: rec.Collection(),
		RKey:       rkey,
		Validate:   validate,
		SwapCommit: (*atproto.CIDAsString)(swap),
	})
}

// PutRecord for the authentificated [Client] (create or update the [Record]).
//
// Rkey used to store and to locate the [Record] (must be set!).
// Validate indicates whether the server should validate the [Record] against the Lexicon schema (if unset, validate
// only for known lexicons).
// SwapCommit is used to compare and swap with the previous commit (not required, disabled by default).
// SwapRecord is used to compare and swap with the previous record (not required, disabled by default ; cannot
// represent nullable easily in Go...).
func PutRecord(
	ctx context.Context,
	client Client,
	rec Record,
	rkey atproto.RecordKey,
	validate *bool,
	swapCommit *atproto.CID,
	swapRecord *atproto.CID,
) (*SendRecordResult, error) {
	req := client.NewRequest().Endpoint(collection.Name("putRecord").Build())
	return sendRecord(ctx, client, req, sendRecordRequest{
		Rec:        rec,
		Repo:       req.GetAuth().DID(),
		Collection: rec.Collection(),
		RKey:       rkey,
		Validate:   validate,
		SwapCommit: (*atproto.CIDAsString)(swapCommit),
		SwapRecord: (*atproto.CIDAsString)(swapRecord),
	})
}

func sendRecord(
	ctx context.Context,
	client Client,
	req RequestBuilder,
	body sendRecordRequest,
) (*SendRecordResult, error) {
	b, err := client.Procedure(ctx, req, AsJsonBodyRequest(body))
	if err != nil {
		return nil, err
	}
	var v SendRecordResult
	return &v, json.Unmarshal(b, &v)
}

// DeleteRecord for the authentificated [Client].
// Doesn't return an error if the [Record] doesn't exist.
//
// Rkey used to locate the [Record] (must be set!).
// SwapCommit is used to compare and swap with the previous commit (not required, disabled by default).
// SwapRecord is used to compare and swap with the previous record (not required, disabled by default ; cannot
// represent nullable easily in Go...).
func DeleteRecord[T Record](
	ctx context.Context,
	client Client,
	rkey atproto.RecordKey,
	swapCommit *atproto.CID,
	swapRecord *atproto.CID,
) (*CommitMeta, error) {
	req := client.NewRequest().Endpoint(collection.Name("deleteRecord").Build())
	var rec T
	b, err := client.Procedure(ctx, req, AsJsonBodyRequest(sendRecordRequest{
		Repo:       req.GetAuth().DID(),
		Collection: rec.Collection(),
		RKey:       rkey,
		SwapCommit: (*atproto.CIDAsString)(swapCommit),
		SwapRecord: (*atproto.CIDAsString)(swapRecord),
	}))
	if err != nil {
		return nil, err
	}
	var v struct {
		Commit CommitMeta `json:"commit"`
	}
	return &v.Commit, json.Unmarshal(b, &v)
}
