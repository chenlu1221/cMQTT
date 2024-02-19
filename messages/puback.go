package messages

import (
	"io"
)

type PuBack struct {
	Header           FixedHeader
	PacketIdentifier uint16 //等待确认的publish的报文标识符
}

func (c *PuBack) Encode(w io.Writer) error {
	return encodePubCommon(w, &c.Header, c.PacketIdentifier)
}

func (c *PuBack) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()
	c.Header = hdr
	//剩余长度，报文第2个字节开始
	remainingLength := decodeLength(r)
	if remainingLength != 2 {
		err = remainingLengthError
	}
	c.PacketIdentifier = getUint16(r, &remainingLength)
	if remainingLength != 0 {
		return msgTooLongError
	}
	return nil
}

// GetPubBack  PUBBacl 报文是对 QoS 等级 2 的 PUBLISH 报文的响应。它是 QoS 2 等级协议交换的第二个报文。
func GetPubBack() *PuBack {
	return &PuBack{
		Header: FixedHeader{MessageType: MsgPubAck},
	}
}
