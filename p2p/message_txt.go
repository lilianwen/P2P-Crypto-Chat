package p2p

import (
	"bytes"
	"encoding/binary"
)

type MsgTxt struct {
	ContentLen uint32
	Content []byte
}

func (t *MsgTxt)Len() int {
	return 4+len(t.Content)
}

func (t *MsgTxt)Serialize() []byte {
	buf := bytes.Buffer{}
	var i2b4 [4]byte
	binary.LittleEndian.PutUint32(i2b4[:], t.ContentLen)
	buf.Write(i2b4[:])
	buf.Write(t.Content[:t.ContentLen])
	return buf.Bytes()
}

func (t *MsgTxt)Parse(data []byte) error {
	if len(data) < 4 {
		return ErrDataTooShort
	}
	t.ContentLen = binary.LittleEndian.Uint32(data[:4])
	if len(data) < int(4+t.ContentLen) {
		return ErrDataTooShort
	}
	t.Content= append(t.Content, data[4:4+t.ContentLen]...)
	return nil
}

func NewMsgTxt(txt string) []byte {
	msg := MsgTxt{
		ContentLen:uint32(len(txt)),
		Content: []byte(txt),
	}
	return msg.Serialize()
}