package research

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"reflect"
	"testing"
)

func TestECDSA(t *testing.T) {
	priv1, pubKey1 := generateKeys(t)
	priv2, pubKey2 := generateKeys(t)

	commonKey1 := btcec.GenerateSharedSecret(priv1, pubKey2)
	commonKey2 := btcec.GenerateSharedSecret(priv2, pubKey1)

	if !reflect.DeepEqual(commonKey1, commonKey2) {
		t.Error("not equal")
		t.Error("commonKey1:", hex.EncodeToString(commonKey1))
		t.Error("commonKey2:", hex.EncodeToString(commonKey2))
	}
}

func generateKeys(t *testing.T) (*btcec.PrivateKey, *btcec.PublicKey) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Error(err)
		return nil, nil
	}

	pubKey := btcec.PublicKey(priv.PublicKey)
	return priv, &pubKey
}
