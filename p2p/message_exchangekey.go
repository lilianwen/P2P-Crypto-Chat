package p2p

type MsgExchangeKey struct {
	PubKey [33]byte
}

func (ek *MsgExchangeKey)Len() int {
	return 33
}

func (ek *MsgExchangeKey)Serialize() []byte {
	return ek.PubKey[:]
}

func (ek *MsgExchangeKey)Parse(data []byte) error {
	if len(data) != ek.Len() {
		return ErrDataLenWrong
	}
	copy(ek.PubKey[:], data[:])
	return nil
}