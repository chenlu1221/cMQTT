package messages

import "io"

type PubRec struct {
	Header           FixedHeader
	PacketIdentifier uint16 //等待确认的publish的报文标识符
}

func (c *PubRec) Encode(w io.Writer) error {
	return encodePubCommon(w, &c.Header, c.PacketIdentifier)
}

func (c *PubRec) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
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

// GetPubRec PUBACK 报文是对 QoS 1 等级的 PUBLISH 报文的响应
func GetPubRec() *PubRec {
	return &PubRec{
		Header: FixedHeader{MessageType: MsgPubRec},
	}
}
