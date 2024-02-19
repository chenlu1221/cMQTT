package messages

import (
	"io"
)

type UnSubAck struct {
	Header           FixedHeader
	PacketIdentifier uint16 //可变报头包含等待确认的UNSUBSCRIBE报文的报文标识符
}

func (c *UnSubAck) Encode(w io.Writer) error {
	return encodePubCommon(w, &c.Header, c.PacketIdentifier)
}

func (c *UnSubAck) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
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

// GetUnSubAck 服务端发送UNSUBACK报文给客户端用于确认收到UNSUBSCRIBE报文。
func GetUnSubAck() *UnSubAck {
	return &UnSubAck{
		Header: FixedHeader{MessageType: MsgUnsubAck},
	}
}
