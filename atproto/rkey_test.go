package atproto

import (
	"testing"

	"pgregory.net/rapid"
)

func genRecordKey() *rapid.Generator[string] {
	return rapid.StringOfN(rapid.RuneFrom(
		[]rune("abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890.-_:~"),
	), 1, -1, 512).Filter(func(s string) bool { return s != "." && s != ".." })
}

var (
	invalidRKey = []string{
		"alpha/beta",
		".",
		"..",
		"#extra",
		"@handle",
		"any space",
		"any+space",
		"number[3]",
		"number(3)",
		`"quote"`,
		"dHJ1ZQ==",
	}
)

func TestParseRKey(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rkey := genRecordKey().Draw(t, "rkey")
		parsed, err := ParseRecordKey(rkey)
		if err != nil {
			t.Fatal(err)
		}
		if parsed.String() != rkey {
			t.Errorf("invalid RecordKey: %s, wanted %s", parsed, rkey)
		}
	})
	for _, rkey := range invalidRKey {
		_, err := ParseRecordKey(rkey)
		if err == nil {
			t.Errorf("expected error for %s", rkey)
		}
	}
}

func TestRecordKey_TID(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tid := genTID(t, "tid")
		rkey, err := ParseRecordKey(tid.String())
		if err != nil {
			t.Fatal(err)
		}
		gotTid, err := rkey.TID()
		if err != nil {
			t.Fatal(err)
		}
		if gotTid != tid {
			t.Errorf("invalid tid: %s, wanted %s", gotTid, tid)
		}
	})
}

func TestRecordKey_NSID(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nsid, rawNsid := genDomain(t, true, "nsid")
		rkey, err := ParseRecordKey(rawNsid)
		if err != nil {
			t.Fatal(err)
		}
		gotNsid, err := rkey.NSID()
		if err != nil {
			t.Fatal(err)
		}
		if *gotNsid != *nsid {
			t.Errorf("invalid nsid: %s, wanted %s", gotNsid, nsid)
		}
	})
}
