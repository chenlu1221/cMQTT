package messages

import (
	"io"
)

type PubRel struct {
	Header           FixedHeader
	PacketIdentifier uint16 //可变报头包含与等待确认的 PUBREC 报文相同的报文标识符。
}

func (c *PubRel) Encode(w io.Writer) error {
	return encodePubCommon(w, &c.Header, c.PacketIdentifier)
}

func (c *PubRel) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
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

// GetPubRel PUBREL 报文是对 PUBREC 报文的响应。它是 QoS 2 等级协议交换的第三个报文。
func GetPubRel() *PubRel {
	return &PubRel{
		Header: FixedHeader{MessageType: MsgPubRel},
	}
}
