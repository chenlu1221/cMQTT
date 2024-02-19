package messages

import (
	"io"
)

type Disconnect struct {
	Header FixedHeader
}

func (d *Disconnect) Encode(w io.Writer) error {
	return d.Header.Encode(w, 0)
}

func (d *Disconnect) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) error {
	d.Header = hdr
	//剩余长度，报文第2个字节开始
	remainingLength := decodeLength(r)
	if remainingLength != 0 {
		return msgTooLongError
	}
	return nil
}

func GetDisconnect() *Disconnect {
	return &Disconnect{
		Header: FixedHeader{MessageType: MsgDisconnect},
	}
}
