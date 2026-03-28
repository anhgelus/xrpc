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
	Value T              `json:"value"`
	URI   atproto.RawURI `json:"uri"`
	CID   string         `json:"cid"`
}

var repoNSID = atproto.NewNSIDBuilder("com.atproto.repo")

// GetRecord returns a single [Record] from a repository.
//
// If cid is omitted, it will return the latest version of the [Record].
func GetRecord[T Record, A atproto.Authority](
	ctx context.Context,
	client Client,
	pds string,
	authority A,
	rkey atproto.RecordKey,
	cid string,
) (RecordStored[T], error) {
	var v RecordStored[T]
	params := make(url.Values)
	params.Add("repo", authority.String())
	params.Add("collection", v.Value.Type().String())
	params.Add("rkey", rkey.String())
	if cid != "" {
		params.Add("cid", cid)
	}
	req := client.NewRequest().
		Server(pds).
		Endpoint(repoNSID.Finish("getRecord")).
		Params(params)
	b, err := client.Query(ctx, req)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(b, &v)
	return v, err
}

type listOut[T Record] struct {
	Cursor  string            `json:"cursor,omitempty"`
	Records []RecordStored[T] `json:"records"`
}

// ListRecords in a repository.
//
// limit is optional (default: 50, max: 100), set to -1 to use default.
// cursor is optional.
func ListRecords[T Record, A atproto.Authority](
	ctx context.Context,
	client Client,
	pds string,
	authority A,
	limit uint8,
	cursor string,
	reverse bool,
) ([]RecordStored[T], string, error) {
	var v T
	params := make(url.Values)
	params.Add("repo", authority.String())
	params.Add("collection", v.Type().String())
	if limit == 0 {
		limit = 50
	}
	params.Add("limit", strconv.Itoa(int(min(limit, 100))))
	if cursor != "" {
		params.Add("cursor", cursor)
	}
	params.Add("reverse", fmt.Sprintf("%t", reverse))
	req := client.NewRequest().
		Server(pds).
		Endpoint(repoNSID.Finish("listRecords")).
		Params(params)
	b, err := client.Query(ctx, req)
	if err != nil {
		return nil, "", err
	}
	var out listOut[T]
	err = json.Unmarshal(b, &out)
	return out.Records, out.Cursor, err
}
