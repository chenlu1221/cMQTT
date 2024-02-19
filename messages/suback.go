package messages

import (
	"bytes"
	"io"
)

type SubAck struct {
	Header           FixedHeader
	PacketIdentifier uint16 //可变报头包含等待确认的SUBSCRIBE报文的报文标识符
	QoSCode          []QoS  //返回码列表
}

func (c *SubAck) Encode(w io.Writer) error {
	buf := new(bytes.Buffer)
	setUint16(c.PacketIdentifier, buf)
	for i := 0; i < len(c.QoSCode); i += 1 {
		setUint8(uint8(c.QoSCode[i]), buf)
	}

	return writeMessage(w, &c.Header, buf)
}

func (c *SubAck) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	c.Header = hdr
	packetRemaining := decodeLength(r)
	c.PacketIdentifier = getUint16(r, &packetRemaining)
	topicsQos := make([]QoS, 0)
	for packetRemaining > 0 {
		grantedQos := QoS(getUint8(r, &packetRemaining))
		if !grantedQos.IsReturnCode() {
			return badReturnCodeError
		}
		topicsQos = append(topicsQos, grantedQos)
	}
	c.QoSCode = topicsQos

	return nil
}

// GetSubAck 服务端发送SUBACK报文给客户端，用于确认它已收到并且正在处理SUBSCRIBE报文。
func GetSubAck() *SubAck {
	return &SubAck{
		Header: FixedHeader{MessageType: MsgSubAck},
	}
}
