package messages

import "io"

type PubComp struct {
	Header           FixedHeader
	PacketIdentifier uint16 //可变报头包含与等待确认的 PUBREL 报文相同的报文标识符。
}

func (c *PubComp) Encode(w io.Writer) error {
	return encodePubCommon(w, &c.Header, c.PacketIdentifier)
}

func (c *PubComp) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
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

// GetPubComp PUBCOMP 报文是对 PUBREL 报文的响应。它是 QoS 2 等级协议交换的第四个也是最后一个报文。
func GetPubComp() *PubComp {
	return &PubComp{
		Header: FixedHeader{MessageType: MsgPubComp},
	}
}
