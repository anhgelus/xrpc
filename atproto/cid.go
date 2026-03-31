package atproto

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// CIDVersion is the supported [CID] version.
const CIDVersion uint = 1

var (
	base32encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)
	base64encoding = base64.StdEncoding.WithPadding(base64.NoPadding)
)

// CIDEncoding describes how [CID] is encoded.
type CIDEncoding interface {
	// Encode a [CID] bytes into a string.
	Encode(*CID) string
	// Decode returns a [CID] from bytes.
	Decode([]byte) (*CID, error)
}

type ErrCannotDecodeAs struct {
	base string
}

func (err ErrCannotDecodeAs) Error() string {
	return "cannot decode string as " + err.base
}

type CIDEncodingBase32 struct{}

func (enc CIDEncodingBase32) Encode(c *CID) string {
	b := c.AsBytes()
	var sb strings.Builder
	sb.Grow(1 + len(b))
	sb.WriteRune('b')
	sb.WriteString(base32encoding.EncodeToString(b))
	return sb.String()
}

func (enc CIDEncodingBase32) Decode(b []byte) (*CID, error) {
	if len(b) < 2 || b[0] != 'b' {
		return nil, ErrCannotDecodeAs{"base32"}
	}
	cp, err := base32encoding.DecodeString(string(b[1:]))
	if err != nil {
		return nil, err
	}
	return ParseCID(cp)
}

var cidEncodings = map[byte]CIDEncoding{'b': CIDEncodingBase32{}}

// CIDCodec is the content codec of the [CID].
type CIDCodec uint

const (
	CIDCodecRaw  CIDCodec = 0x55
	CIDCodecCBOR CIDCodec = 0x71
)

var cidCodecs = []CIDCodec{CIDCodecRaw, CIDCodecCBOR}

// CIDHashType indicates the hash function used in the [CID].
type CIDHashType uint

const (
	CIDHashSha256 CIDHashType = 0x12
)

var cidHashes = map[CIDHashType]uint{
	CIDHashSha256: 32,
}

// CID is a DASL Content ID.
//
// See [ParseCID] to parse it from bytes.
// See [CIDAsString] to marshal/unmarshal as string.
// See [CIDLink] to marshal/unmarshal as cid-link.
type CID struct {
	Version  uint
	Codec    CIDCodec
	HashType CIDHashType
	HashSize uint
	Digest   []byte
}

var (
	ErrNotCID                 = errors.New("not a CID")
	ErrUnsupportedCIDVersion  = errors.New("unsupported CID version")
	ErrUnsupportedCIDCodec    = errors.New("unsupported CID codec")
	ErrUnsupportedCIDHash     = errors.New("unsupported CID hash")
	ErrUnsupportedCIDEncoding = errors.New("unsupported CID encoding")
)

// ParseCID from bytes.
//
// See [ParseCIDString] to extract it from a string.
func ParseCID(b []byte) (*CID, error) {
	fmt.Printf("%v\n%d\n", b, len(b))
	if len(b) < 5 {
		return nil, ErrNotCID
	}
	var c CID
	c.Version = uint(b[0])
	if c.Version != CIDVersion {
		return nil, ErrUnsupportedCIDVersion
	}
	c.Codec = CIDCodec(b[1])
	if !slices.Contains(cidCodecs, c.Codec) {
		return nil, ErrUnsupportedCIDCodec
	}
	c.HashType = CIDHashType(b[2])
	size, ok := cidHashes[c.HashType]
	if !ok {
		return nil, ErrUnsupportedCIDHash
	}
	c.HashSize = uint(b[3])
	next := b[4:]
	if len(next) != int(c.HashSize) || size != c.HashSize {
		return nil, ErrNotCID
	}
	c.Digest = next
	return &c, nil
}

// AsBytes returns bytes representation of the [CID].
func (c *CID) AsBytes() []byte {
	b := make([]byte, 4+c.HashSize)
	b[0] = byte(c.Version)
	b[1] = byte(c.Codec)
	b[2] = byte(c.HashType)
	b[3] = byte(c.HashSize)
	for i, v := range c.Digest {
		b[i+4] = v
	}
	return b
}

// AsString returns the string representation of the [CID] using the specified encoder.
func (c *CID) AsString(enc CIDEncoding) string {
	return enc.Encode(c)
}

// String is [CID.AsString] with the default encoding ([CIDEncodingBase32]).
func (c *CID) String() string {
	return c.AsString(CIDEncodingBase32{})
}

// CIDAsString marshals and unmarshals a [CID] as a string with the default encoding.
//
// Used as ATProto string field type with format cid.
//
// See [CID.String] for more information.
type CIDAsString CID

func (c *CIDAsString) CID() *CID {
	return (*CID)(c)
}

func (c *CIDAsString) String() string {
	return c.CID().String()
}

func (c *CIDAsString) MarshalMap() (any, error) {
	return c.CID().String(), nil
}

func (c *CIDAsString) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	enc, ok := cidEncodings[s[0]]
	if !ok {
		return ErrUnsupportedCIDEncoding
	}
	cid, err := enc.Decode([]byte(s))
	if err != nil {
		return err
	}
	*c = CIDAsString(*cid)
	return nil
}

// CIDLink marshals and unmarshals a [CID] as ATProto Lexicon type cid-link.
type CIDLink CID

func (c *CIDLink) CID() *CID {
	return (*CID)(c)
}

func (c *CIDLink) String() string {
	return c.CID().String()
}

func (c *CIDLink) MarshalMap() (any, error) {
	cid := CIDAsString(*c)
	v, err := cid.MarshalMap()
	return map[string]any{"$link": v}, err
}

func (c *CIDLink) UnmarshalJSON(b []byte) error {
	var v struct {
		Link *CIDAsString `json:"$link"`
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	*c = CIDLink(*v.Link)
	return nil
}

// CIDBytes marshals and unmarshals a [CID] like [CIDLink], but with $bytes instead of $link.
type CIDBytes CID

func (c *CIDBytes) CID() *CID {
	return (*CID)(c)
}

func (c *CIDBytes) String() string {
	return c.CID().String()
}

func (c *CIDBytes) MarshalMap() (any, error) {
	return map[string]any{
		"$bytes": base64encoding.EncodeToString(c.CID().AsBytes()),
	}, nil
}

func (c *CIDBytes) UnmarshalJSON(b []byte) error {
	var v struct {
		Bytes string `json:"$bytes"`
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	raw, err := base64encoding.DecodeString(v.Bytes)
	if err != nil {
		return err
	}
	cid, err := ParseCID(raw)
	if err != nil {
		return err
	}
	*c = CIDBytes(*cid)
	return nil
}
