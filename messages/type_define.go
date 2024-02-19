package messages

import (
	"bytes"
	"io"
)

type returnCode uint8
type QoS uint8

func (q QoS) IsReturnCode() bool {
	return q&0x7c == 0
}

// HasId 等于1或2
func (q QoS) HasId() bool {
	return q == QosAtLeastOnce || q == QosExactlyOnce
}
func (q QoS) IsValid() bool {
	return q < qosFirstInvalid
}

const (
	RetCodeAccepted                    = returnCode(iota) //连接已被服务端接受
	RetCodeUnacceptableProtocolVersion                    //服务端不支持客户端请求的 MQTT 协议级别
	RetCodeIdentifierRejected                             //客户端标识符是正确的 UTF-8 编码，但服务 端不允许使用
	RetCodeServerUnavailable                              //网络连接已建立，但 MQTT 服务不可用
	RetCodeBadUsernameOrPassword                          // 用户名或密码的数据格式无效
	RetCodeNotAuthorized                                  //客户端未被授权连接到此服务器
	retCodeFirstInvalid                                   //保留

	// MaxPayloadSize 256MB-1B
	MaxPayloadSize = (1 << (4 * 7)) - 1
)

func (rc returnCode) IsValid() bool {
	return rc < retCodeFirstInvalid
}

const (
	QosAtMostOnce = QoS(iota)
	QosAtLeastOnce
	QosExactlyOnce

	qosFirstInvalid
	qosFailure = 8
)

type QoSCode uint8
type topics struct {
	Topic string
	QoS   QoS
}
type DefaultDecoderConfig struct{}

func (c DefaultDecoderConfig) MakePayload(msg *Publish, r io.Reader, n int) (PayloadIntFace, error) {
	return make(BytesPayload, n), nil
}

// DecodeOneMessage 从io中解码一条消息
func DecodeOneMessage(r io.Reader, config DecoderConfig) (msg Message, err error) {
	var hdr FixedHeader
	err = hdr.Decode(r)
	if err != nil {
		return
	}

	msg, err = NewMessage(hdr.MessageType)
	if err != nil {
		return
	}

	if config == nil {
		config = DefaultDecoderConfig{}
	}

	return msg, msg.Decode(r, hdr, config)
}
func NewMessage(msgType messageType) (msg Message, err error) {
	switch msgType {
	case MsgConnect:
		msg = GetConnect()
	case MsgConnAck:
		msg = GetConNack()
	case MsgPublish:
		msg = GetPublish()
	case MsgPubAck:
		msg = GetPubBack()
	case MsgPubRec:
		msg = GetPubRec()
	case MsgPubRel:
		msg = GetPubRel()
	case MsgPubComp:
		msg = GetPubComp()
	case MsgSubscribe:
		msg = GetSubScribe()
	case MsgUnsubAck:
		msg = GetUnSubAck()
	case MsgSubAck:
		msg = GetSubAck()
	case MsgUnsubscribe:
		msg = GetUnSubScribe()
	case MsgPingReq:
		msg = GetPingReq()
	case MsgPingResp:
		msg = GetPingResp()
	case MsgDisconnect:
		msg = GetDisconnect()
	default:
		return nil, badMsgTypeError
	}

	return
}

func setUint16(val uint16, buf *bytes.Buffer) {
	buf.WriteByte(byte(val & 0xff00 >> 8))
	buf.WriteByte(byte(val & 0x00ff))
}

func setString(val string, buf *bytes.Buffer) {
	length := uint16(len(val))
	setUint16(length, buf)
	buf.WriteString(val)
}
func boolToByte(val bool) byte {
	if val {
		return byte(1)
	}
	return byte(0)
}
func writeMessage(w io.Writer, hdr *FixedHeader, payloadBuf *bytes.Buffer) error {
	totalPayloadLength := int64(len(payloadBuf.Bytes()))
	if totalPayloadLength > MaxPayloadSize {
		return msgTooLongError
	}

	buf := new(bytes.Buffer)
	//编码固定头
	err := hdr.encodeInto(buf, int32(totalPayloadLength))
	if err != nil {
		return err
	}

	//写入剩余内容
	buf.Write(payloadBuf.Bytes())
	_, err = w.Write(buf.Bytes())

	return err
}

// 固定头的剩余长度字段
func encodeLength(length int32, buf *bytes.Buffer) {
	if length == 0 {
		buf.WriteByte(0)
		return
	}
	//变长编码
	for length > 0 {
		digit := length & 0x7f
		length = length >> 7
		if length > 0 {
			digit = digit | 0x80
		}
		buf.WriteByte(byte(digit))
	}
}
func encodePubCommon(w io.Writer, hdr *FixedHeader, packetIdentifier uint16) error {
	buf := new(bytes.Buffer)
	setUint16(packetIdentifier, buf)
	return writeMessage(w, hdr, buf)
}
func setUint8(u uint8, buf *bytes.Buffer) {
	buf.WriteByte(byte(u))
}
