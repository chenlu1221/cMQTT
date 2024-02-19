package messages

import (
	"bytes"
	"io"
)

type ConNack struct {
	Header         FixedHeader
	SessionPresent bool
	ReturnCode     returnCode
}

func (c *ConNack) Encode(w io.Writer) error {
	buf := new(bytes.Buffer)

	flags := boolToByte(c.SessionPresent)
	buf.WriteByte(flags)
	buf.WriteByte(uint8(c.ReturnCode))

	return writeMessage(w, &c.Header, buf)
}

func (c *ConNack) Decode(r io.Reader, hdr FixedHeader, config DecoderConfig) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()
	c.Header = hdr
	//剩余长度，报文第2个字节开始
	remainingLength := decodeLength(r)
	if remainingLength != 2 {
		err = remainingLengthError
	}
	//解码SessionPresent字段，判定是否合规
	sp := getUint8(r, &remainingLength)
	if sp > 1 {
		return sessionPresentError
	}
	c.SessionPresent = sp&0x01 > 0
	c.ReturnCode = returnCode(getUint8(r, &remainingLength))
	if !c.ReturnCode.IsValid() {
		return badReturnCodeError
	}

	if remainingLength != 0 {
		return msgTooLongError
	}
	return nil
}

func GetConNack() *ConNack {
	return &ConNack{
		Header: FixedHeader{MessageType: MsgConnAck},
	}
}
