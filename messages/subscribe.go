package messages

import (
	"bytes"
	"io"
)

type SubScribe struct {
	Header           FixedHeader
	PacketIdentifier uint16   //客户端标识符
	TopicList        []topics //主题列表
}

func (c *SubScribe) Encode(w io.Writer) error {
	buf := new(bytes.Buffer)
	if c.Header.QoS.HasId() {
		setUint16(c.PacketIdentifier, buf)
	}
	for _, topicSub := range c.TopicList {
		setString(topicSub.Topic, buf)
		setUint8(uint8(topicSub.QoS), buf)
	}

	return writeMessage(w, &c.Header, buf)
}

func (c *SubScribe) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	c.Header = hdr
	packetRemaining := decodeLength(r)
	if c.Header.QoS.HasId() {
		c.PacketIdentifier = getUint16(r, &packetRemaining)
	}
	var tos []topics
	for packetRemaining > 0 {
		tos = append(tos, topics{
			Topic: getString(r, &packetRemaining),
			QoS:   QoS(getUint8(r, &packetRemaining)),
		})
	}
	c.TopicList = tos

	return nil
}

// GetSubScribe 客户端向服务端发送SUBSCRIBE报文用于创建一个或多个订阅。
func GetSubScribe() *SubScribe {
	return &SubScribe{
		Header: FixedHeader{MessageType: MsgSubscribe},
	}
}
