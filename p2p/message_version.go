package p2p

import (
	"bytes"
	"encoding/binary"
)

const (
	CryptoSuite = "AES_128_SECP256K1"
)

type MsgVersion struct {
	Version uint32
	CryptoSuit [20]byte
}

func (v *MsgVersion)Len() int {
	return 24
}

func (v *MsgVersion)Serialize() []byte{
	buf := bytes.Buffer{}
	var i2b4 [4]byte
	binary.LittleEndian.PutUint32(i2b4[:], v.Version)
	buf.Write(i2b4[:])
	buf.Write(v.CryptoSuit[:])
	return buf.Bytes()
}

func (v *MsgVersion)Parse(data []byte) error {
	if len(data) != v.Len() {
		return ErrDataLenWrong
	}
	v.Version = binary.LittleEndian.Uint32(data[:4])
	copy(v.CryptoSuit[:], data[4:24])
	return nil
}

func NewVerMsg() []byte {
	msg := MsgVersion{
		Version:uint32(0x10000000),
	}
	copy(msg.CryptoSuit[:], []byte(CryptoSuite))
	return msg.Serialize()
}