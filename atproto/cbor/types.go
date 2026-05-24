package cbor

type majorType byte

const (
	unsignedInt  majorType = 0b000
	negativeInt  majorType = 0b001
	byteString   majorType = 0b010
	textString   majorType = 0b011
	array        majorType = 0b100
	mapT         majorType = 0b101
	tag          majorType = 0b110
	simpleValues majorType = 0b111
)

type additionalType byte

const (
	nextUint8        additionalType = 0b11000
	nextUint16       additionalType = 0b11001
	nextUint32       additionalType = 0b11010
	nextUint64       additionalType = 0b11011
	indefiniteLength additionalType = 0b11111
)

type Tag interface {
	Tag() uint64
	Value() any
}

func MarshalTag(t Tag) ([]byte, error) {
	in, err := Marshal(t.Value())
	if err != nil {
		return nil, err
	}
	b := marshalRawInt(tag, t.Tag())
	return append(b, in...), nil
}
