package messages

import (
	"bytes"
	"io"
)

// MessageType constants.
const (
	MsgConnect = messageType(iota + 1)
	MsgConnAck
	MsgPublish
	MsgPubAck
	MsgPubRec
	MsgPubRel
	MsgPubComp
	MsgSubscribe
	MsgSubAck
	MsgUnsubscribe
	MsgUnsubAck
	MsgPingReq
	MsgPingResp
	MsgDisconnect

	msgTypeFirstInvalid
)

type messageType uint8

func (mt messageType) IsValid() bool {
	return mt >= MsgConnect && mt < msgTypeFirstInvalid
}

// FixedHeader 固定报头
type FixedHeader struct {
	MessageType messageType
	Dup         bool //如果DUP标志被设置为0，表示这是客户端或服务端第一次请求发送这个PUBLISH报文。如果DUP标志被设置为1，表示这可能是一个早前报文请求的重发。
	QoS         QoS  //质量服务等级
	Retain      bool //保留标志
}

// 编码固定头，写入buf
func (hdr *FixedHeader) encodeInto(buf *bytes.Buffer, remainingLength int32) error {
	if !hdr.QoS.IsValid() {
		return badQosError
	}
	if !hdr.MessageType.IsValid() {
		return badMsgTypeError
	}

	val := byte(hdr.MessageType) << 4
	val |= boolToByte(hdr.Dup) << 3
	val |= byte(hdr.QoS) << 1
	val |= boolToByte(hdr.Retain)
	buf.WriteByte(val)
	encodeLength(remainingLength, buf)
	return nil
}
func (hdr *FixedHeader) Encode(w io.Writer, remainingLength int32) error {
	buf := new(bytes.Buffer)
	err := hdr.encodeInto(buf, remainingLength)
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}

// Decode 从io中解析出固定头
func (hdr *FixedHeader) Decode(r io.Reader) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	var buf [1]byte
	if _, err = io.ReadFull(r, buf[:]); err != nil {
		return
	}
	byte1 := buf[0]
	msgType := messageType(byte1 >> 4)

	*hdr = FixedHeader{
		MessageType: msgType,
		Dup:         byte1&0x08 > 0,
		QoS:         QoS(byte1 & 0x06 >> 1),
		Retain:      byte1&0x01 > 0,
	}
	return
}

// 解码出变长编码的剩余长度字段
func decodeLength(r io.Reader) int32 {
	var v int32
	var buf [1]byte
	var shift uint
	for i := 0; i < 4; i++ {
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			raiseError(err)
		}

		b := buf[0]
		v |= int32(b&0x7f) << shift

		if b&0x80 == 0 {
			return v
		}
		shift += 7
	}

	raiseError(badLengthEncodingError)
	panic("unreachable")
}
