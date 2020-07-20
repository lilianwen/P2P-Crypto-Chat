package research

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"
)

func TestAes(t *testing.T) {
	key := []byte("deadbeefdeadbeef")
	blk, err := aes.NewCipher(key[:])
	if err != nil {
		t.Error(err)
		return
	}

	aead, err := cipher.NewGCM(blk)
	if err != nil {
		t.Error(err)
		return
	}

	nonce := make([]byte, 12)
	if false {
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			panic(err.Error())
		}
	}

	var encryptedMsg []byte
	var decryptedMsg []byte
	message := []byte("i love you!!")
	encryptedMsg = aead.Seal(nil, nonce, message, nil)
	t.Log(hex.EncodeToString(encryptedMsg))

	if decryptedMsg, err = aead.Open(nil, nonce, encryptedMsg, nil ); err != nil {
		t.Error(err)
		return
	}
	t.Log(string(decryptedMsg))

}
