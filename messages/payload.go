package messages

import (
	"bytes"
	"io"
)

// ConnectPayload  报文的有效载荷（payload）包含一个或多个以长度为前缀的字段，可变报头中的标志决定是否
// 包含这些字段。如果包含的话，必须按这个顺序出现：客户端标识符，遗嘱主题，遗嘱消息，用户名，密
// 码
type ConnectPayload struct {
	ClientId    string
	WillTopic   string //遗嘱主题
	WillMessage string
	UserName    string
	Password    string
}
type PubilshPayload struct {
	ClientId    string
	WillTopic   string //遗嘱主题
	WillMessage string
	UserName    string
	Password    string
}
type BytesPayload []byte

func (p BytesPayload) Size() int {
	return len(p)
}

func (p BytesPayload) WritePayload(w io.Writer, buf *bytes.Buffer) error {
	_, err := buf.Write(p)
	return err
}

func (p BytesPayload) ReadPayload(r io.Reader) error {
	_, err := io.ReadFull(r, p)
	return err
}
