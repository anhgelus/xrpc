package atproto

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

var (
	invalidNSID = []string{
		"com.exa💩ple.thing",
		"com.example",
		"com.example.3",
	}
)

func genDomain(t *rapid.T, isNsid bool, label string) (*NSID, string) {
	ascii := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	asciiNums := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	asciiNumsHyp := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"))
	authority := func(l string, gen *rapid.Generator[rune]) string {
		a := rapid.StringOfN(ascii, 1, -1, 1).Draw(t, l+" first")
		if rapid.Bool().Draw(t, l+" bigger") {
			a += rapid.StringOfN(gen, 0, -1, 61).Draw(t, l+" second") +
				rapid.StringOfN(asciiNums, 1, -1, 1).Draw(t, l+" final")
		}
		return strings.ToLower(a)
	}
	var sb strings.Builder
	if isNsid {
		sb.WriteString(authority(label+" tld", asciiNumsHyp))
	}
	ok := true
	for i := 0; ok; i++ {
		s := authority(fmt.Sprintf("%s sub %d", label, i), asciiNumsHyp)
		if sb.Len()+len(s)+1 <= NSIDMaxLength {
			if isNsid {
				sb.WriteRune('.')
			}
			sb.WriteString(s)
			if !isNsid {
				sb.WriteRune('.')
			}
			ok = rapid.Bool().Draw(t, label+" continue?")
		} else {
			ok = false
		}
	}
	if !isNsid {
		sb.WriteString(authority(label+" tld", asciiNumsHyp))
	}
	var nsid NSID
	nsid.Authority = strings.ToLower(sb.String())
	if isNsid {
		nsid.Name = authority(label+" name", asciiNums)
	}
	return &nsid, nsid.String()
}

func TestParseNSID(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nsid, raw := genDomain(t, true, "nsid")
		t.Log(raw)
		n, err := ParseNSID(raw)
		if err != nil {
			t.Fatal(err)
		}
		if n.Authority != nsid.Authority {
			t.Errorf("invalid authority: %s, wanted %s", n.Authority, nsid.Authority)
		}
		if n.Name != nsid.Name {
			t.Errorf("invalid name: %s, wanted %s", n.Name, nsid.Name)
		}
	})
	for _, nsid := range invalidNSID {
		_, err := ParseNSID(nsid)
		if err == nil {
			t.Errorf("expected error for %s", nsid)
		}
	}
}

func TestNSIDBuilder(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tt1, _ := genDomain(t, true, "base")
		base := NewNSIDBuilder(tt1.Authority)
		tt2, _ := genDomain(t, true, "next")
		baseNext := base.Add(tt2.Authority)
		if base.String() == baseNext.String() {
			t.Errorf("base is baseNext: %s, %s", base.String(), baseNext.String())
		}
		if base.authority != tt1.Authority {
			t.Errorf("invalid authority for base: %s, wanted %s", base.authority, tt1.Authority)
		}
		if baseNext.authority != tt1.Authority+"."+tt2.Authority {
			t.Errorf("invalid authority for baseNext: %s, wanted %s", baseNext.authority, tt1.Authority+"."+tt2.Authority)
		}
		baseNsid := base.Finish(tt1.Name)
		if *baseNsid != *tt1 {
			t.Errorf("invalid nsid finish: %s, wanted %s", baseNsid, tt1)
		}
	})
}
