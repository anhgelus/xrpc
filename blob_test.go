package xrpc

import (
	"encoding/json"
	"testing"

	"pgregory.net/rapid"
)

func genBlob(t *rapid.T, baseMime, label string) (*Blob, map[string]any) {
	blob := &Blob{
		CID: rapid.StringN(2, -1, 128).Draw(t, label+"_cid"),
		MimeType: baseMime + "/" +
			rapid.StringOfN(rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyz")), 2, 20, -1).Draw(t, label+"_mimeType"),
		Size: rapid.UintMin(1).Draw(t, label+"_size"),
	}
	return blob, map[string]any{
		"$type":    blob.Collection(),
		"ref":      map[string]any{"$link": blob.CID},
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
		if bl.CID != blob.CID {
			t.Errorf("invalid CID: %s, wanted %s", bl.CID, blob.CID)
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
		t.Log(string(b))
	})
}
