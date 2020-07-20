package p2p

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ErrDataTooShort  = errors.New("data too short")
	ErrDataLenWrong  = errors.New("data len wrong")
	ErrMsgCmdTooLong = errors.New("cmd of message too long")

)

const (
	Magic = uint32(0x10000000)
	HeadLen = 20
	CmdMaxLen = 12
)

type Header struct {
	Magic uint32
	Command [CmdMaxLen]byte
	PayloadLen uint32
}

func (h *Header)Serialize() []byte {
	buf := bytes.Buffer{}
	var i2b4 [4]byte
	binary.LittleEndian.PutUint32(i2b4[:], h.Magic)
	buf.Write(i2b4[:])
	buf.Write(h.Command[:])
	binary.LittleEndian.PutUint32(i2b4[:], h.PayloadLen)
	buf.Write(i2b4[:])

	return buf.Bytes()
}

func (h *Header)Parse(data []byte) error {
	if len(data) != h.Len() {
		return ErrDataLenWrong
	}

	h.Magic = binary.LittleEndian.Uint32(data[:4])
	copy(h.Command[:], data[4:16])
	h.PayloadLen = binary.LittleEndian.Uint32(data[16:20])
	return nil
}

func (h *Header)Len() int {
	return HeadLen
}

func newHeader(cmd string) Header {
	if len(cmd) > CmdMaxLen {
		panic(ErrMsgCmdTooLong)
	}
	header := Header{
		Magic:Magic,
		PayloadLen:0,
	}
	copy(header.Command[:], []byte(cmd))
	return header
}

func NewMsg(cmd string, payload []byte) []byte {
	h := newHeader(cmd)
	h.PayloadLen = uint32(len(payload))
	var buf = bytes.Buffer{}
	buf.Write(h.Serialize())
	buf.Write(payload)
	return buf.Bytes()
}

