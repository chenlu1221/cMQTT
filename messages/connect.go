package messages

import (
	"bytes"
	"io"
)

type Connect struct {
	Header        FixedHeader
	ProtocolName  string
	ProtocolLevel uint8
	KeepAlive     uint16
	ConnectFlags  uint8
	Payload       ConnectPayload
}

func (c *Connect) Encode(w io.Writer) error {
	if c.ConnectFlags&0x18>>3 >= 3 {
		return badWillQosError
	}

	buf := new(bytes.Buffer)

	var flags = c.ConnectFlags

	setString(c.ProtocolName, buf)
	setUint8(c.ProtocolLevel, buf)
	buf.WriteByte(flags)
	setUint16(c.KeepAlive, buf)
	setString(c.Payload.ClientId, buf)
	if c.ConnectFlags&0x04>>2 > 0 {
		setString(c.Payload.WillTopic, buf)
		setString(c.Payload.WillMessage, buf)
	}
	if c.ConnectFlags&0x80>>7 > 0 {
		setString(c.Payload.UserName, buf)
	}
	if c.ConnectFlags&0x40>>6 > 0 {
		setString(c.Payload.Password, buf)
	}

	return writeMessage(w, &c.Header, buf)
}

func (c *Connect) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	c.Header = hdr
	//剩余长度，报文第2个字节开始
	remainingLength := decodeLength(r)

	protocolName := getString(r, &remainingLength)
	protocolLevel := getUint8(r, &remainingLength)
	ConnectFlags := getUint8(r, &remainingLength)
	keepAliveTimer := getUint16(r, &remainingLength)
	clientId := getString(r, &remainingLength)

	c.ProtocolName = protocolName
	c.ProtocolLevel = protocolLevel
	c.ConnectFlags = ConnectFlags
	c.KeepAlive = keepAliveTimer
	c.Payload.ClientId = clientId

	if c.ConnectFlags&0x04 > 0 {
		c.Payload.WillTopic = getString(r, &remainingLength)
		c.Payload.WillMessage = getString(r, &remainingLength)
	}
	if c.ConnectFlags&0x80 > 0 {
		c.Payload.UserName = getString(r, &remainingLength)
	}
	if c.ConnectFlags&0x40 > 0 {
		c.Payload.Password = getString(r, &remainingLength)
	}

	if remainingLength != 0 {
		return msgTooLongError
	}

	return nil
}

func GetConnect() *Connect {
	return &Connect{
		Header: FixedHeader{MessageType: MsgConnect},
	}
}
