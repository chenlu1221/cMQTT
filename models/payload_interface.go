package models

import (
	"bytes"
	"fmt"
	"io"
)

type intPayload string

func newIntPayload(i int64) intPayload {
	return intPayload(fmt.Sprint(i))
}
func (ip intPayload) ReadPayload(r io.Reader) error {
	// not implemented
	return nil
}
func (ip intPayload) WritePayload(w io.Writer, buf *bytes.Buffer) error {
	_, err := w.Write([]byte(string(ip)))
	return err
}
func (ip intPayload) Size() int {
	return len(ip)
}
