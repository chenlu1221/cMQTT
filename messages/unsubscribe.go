package messages

import (
	"bytes"
	"io"
)

type UnSubScribe struct {
	Header           FixedHeader
	PacketIdentifier uint16   //客户端标识符
	TopicList        []string //想要取消的主题列表
}

func (c *UnSubScribe) Encode(w io.Writer) error {
	buf := new(bytes.Buffer)
	setUint16(c.PacketIdentifier, buf)
	for _, topic := range c.TopicList {
		setString(topic, buf)
	}

	return writeMessage(w, &c.Header, buf)
}

func (c *UnSubScribe) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	c.Header = hdr
	if c.Header.QoS != 1 || c.Header.Dup || c.Header.Retain {
		return badQosError
	}
	packetRemaining := decodeLength(r)
	c.PacketIdentifier = getUint16(r, &packetRemaining)
	tos := make([]string, 0)
	for packetRemaining > 0 {
		tos = append(tos, getString(r, &packetRemaining))
	}
	if c.TopicList = tos; len(c.TopicList) == 0 {
		return badPayloadError
	}

	return nil
}

// GetUnSubScribe 客户端发送UNSUBSCRIBE报文给服务端，用于取消订阅主题。
func GetUnSubScribe() *UnSubScribe {
	return &UnSubScribe{
		Header: FixedHeader{MessageType: MsgUnsubscribe},
	}
}
