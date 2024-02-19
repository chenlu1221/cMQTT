package messages

import (
	"bytes"
	"io"
)

type PayloadIntFace interface {
	Size() int
	WritePayload(w io.Writer, buf *bytes.Buffer) error
	ReadPayload(r io.Reader) error
}
type Message interface {
	Encode(w io.Writer) error
	Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) error
}
type DecoderConfig interface {
	MakePayload(msg *Publish, r io.Reader, n int) (PayloadIntFace, error)
}
