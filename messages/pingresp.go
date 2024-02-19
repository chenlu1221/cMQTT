package messages

import "io"

type PingReq struct {
	Header FixedHeader
}

func (c *PingReq) Encode(w io.Writer) error {
	return c.Header.Encode(w, 0)
}

func (c *PingReq) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	c.Header = hdr
	//剩余长度，报文第2个字节开始
	remainingLength := decodeLength(r)
	if remainingLength != 0 {
		return msgTooLongError
	}
	return nil
}

func GetPingReq() *PingReq {
	return &PingReq{
		Header: FixedHeader{MessageType: MsgPingReq},
	}
}
