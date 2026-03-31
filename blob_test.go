package xrpc

import (
	"crypto/sha256"
	"encoding/json"
	"slices"
	"testing"

	"pgregory.net/rapid"
	"tangled.org/anhgelus.world/xrpc/atproto"
)

func genCID(t *rapid.T, label string) *atproto.CID {
	cid := &atproto.CID{
		Version:  atproto.CIDVersion,
		Codec:    atproto.CIDCodecRaw,
		HashType: atproto.CIDHashSha256,
		HashSize: 32,
	}
	str := rapid.StringN(64, -1, -1).Draw(t, label)
	cp := make([]byte, 32)
	for i, v := range sha256.Sum256([]byte(str)) {
		cp[i] = v
	}
	cid.Digest = cp
	return cid
}

func genBlob(t *rapid.T, baseMime, label string) (*Blob, map[string]any) {
	blob := &Blob{
		CID: (*atproto.CIDLink)(genCID(t, label+"_cid")),
		MimeType: baseMime + "/" +
			rapid.StringOfN(rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyz")), 2, 20, -1).Draw(t, label+"_mimeType"),
		Size: rapid.UintMin(1).Draw(t, label+"_size"),
	}
	v, err := MarshalToMap(blob.CID)
	if err != nil {
		panic(err)
	}
	return blob, map[string]any{
		"$type":    blob.Collection(),
		"ref":      v,
		"mimeType": blob.MimeType,
		"size":     blob.Size,
	}
}

func TestBlob_JSON(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		base := rapid.StringN(1, -1, 10).Draw(t, "mimeType")
		blob, blobRaw := genBlob(t, base, "blob")
		b, err := json.Marshal(blobRaw)
		if err != nil {
			t.Fatal(err)
		}
		var bl *Blob
		err = json.Unmarshal(b, &bl)
		if err != nil {
			t.Fatal(err)
		}
		blCID := bl.CID.CID()
		blobCID := blob.CID.CID()
		if blCID.Version != blobCID.Version {
			t.Errorf("invalid CID version: %d, wanted %d", blCID.Version, blobCID.Version)
		}
		if blCID.Codec != blobCID.Codec {
			t.Errorf("invalid CID codec: %d, wanted %d", blCID.Codec, blobCID.Codec)
		}
		if blCID.HashType != blobCID.HashType {
			t.Errorf("invalid CID hash size: %d, wanted %d", blCID.HashType, blobCID.HashType)
		}
		if blCID.HashSize != blobCID.HashSize {
			t.Errorf("invalid CID hash size: %d, wanted %d", blCID.HashSize, blobCID.HashSize)
		}
		if !slices.Equal(blCID.Digest, blobCID.Digest) {
			t.Errorf("invalid CID digest: %v, wanted %v", blCID.Digest, blobCID.Digest)
		}
		if bl.MimeType != blob.MimeType {
			t.Errorf("invalid mimeType: %s, wanted %s", bl.MimeType, blob.MimeType)
		}
		if bl.Size != blob.Size {
			t.Errorf("invalid size: %d, wanted %d", bl.Size, blob.Size)
		}
		b, err = json.Marshal(bl)
		if err != nil {
			t.Fatal(err)
		}
	})
}
