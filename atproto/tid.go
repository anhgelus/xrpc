package atproto

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	Base32SortAlphabet = "234567abcdefghijklmnopqrstuvwxyz"
	TIDLen             = 13

	clockIdBits = 0x3FF
)

var regexpTID = regexp.MustCompile(`^[234567abcdefghij][234567abcdefghijklmnopqrstuvwxyz]{12}$`)

var ErrInvalidTID = errors.New("invalid TID")

// TID represents a timestamp identifier.
//
// See [ParseTID] to parse a [TID] from a string.
// See [TIDGenerator] to generate [TID]s.
type TID string

// ParseTID in the raw given string.
//
// Returns [ErrInvalidTID] if the [TID] is invalid.
//
// See [TIDGenerator] to generate [TID]s.
func ParseTID(raw string) (TID, error) {
	if !regexpTID.MatchString(raw) {
		return "", ErrInvalidTID
	}
	return TID(raw), nil
}

func newTIDFromInteger(t int64, clockId uint) TID {
	if clockId > clockIdBits {
		panic("invalid clock id")
	}
	v := (uint64(t&0x1F_FFFF_FFFF_FFFF) << 10) | uint64(clockId&clockIdBits)
	v = (0x7FFF_FFFF_FFFF_FFFF & v)
	var sb strings.Builder
	sb.Grow(TIDLen)
	for range TIDLen {
		sb.WriteString(string(Base32SortAlphabet[v&0x1F]))
		v = v >> 5
	}
	// must reverse because it is big-endian
	r := make([]rune, TIDLen)
	for i, v := range []rune(sb.String()) {
		r[TIDLen-i-1] = v
	}
	return TID(string(r))
}

// NewTID from a UNIX timestamp (in milliseconds) and clock ID.
//
// Prefer using [TIDGenerator].
func NewTID(t time.Time, clockId uint) TID {
	return newTIDFromInteger(t.UnixMicro(), clockId)
}

// NewTIDNow with the given clock ID.
//
// Prefer using [TIDGenerator].
func NewTIDNow(clockId uint) TID {
	return NewTID(time.Now(), clockId)
}

func (t TID) String() string {
	return string(t)
}

func (t TID) Uint64() uint64 {
	s := t.String()
	var v uint64
	for i := range TIDLen {
		c := strings.IndexByte(Base32SortAlphabet, s[i])
		if c < 0 {
			return 0
		}
		v = (v << 5) | uint64(c&0x1F)
	}
	return v
}

func (t TID) Time() time.Time {
	i := t.Uint64()
	i = (i >> 10) & 0x1FFF_FFFF_FFFF_FFFF
	return time.UnixMicro(int64(i)).UTC()
}

// ClockID returns the clock ID used to create the [TID].
func (t TID) ClockID() uint {
	i := t.Uint64()
	return uint(i & clockIdBits)
}

func (t TID) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t TID) MarshalMap() (any, error) {
	return t.String(), nil
}

func (t *TID) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*t, err = ParseTID(s)
	return err
}

// TIDGenerator generates [TID] that always increase.
//
// See [NewTIDGenerator].
type TIDGenerator struct {
	clock    uint
	mu       sync.Mutex
	lastUnix int64
}

// NewTIDGenerator using clockID as base clock.
func NewTIDGenerator(clockId uint) *TIDGenerator {
	if clockId > clockIdBits {
		panic("invalid clock id")
	}
	return &TIDGenerator{clock: clockId}
}

// Next returns a new [TID].
func (g *TIDGenerator) Next() TID {
	now := time.Now()
	nUnix := now.Unix()

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lastUnix >= nUnix {
		g.lastUnix += 1
	} else {
		g.lastUnix = nUnix
	}
	return newTIDFromInteger(g.lastUnix, g.clock)
}
