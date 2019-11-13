package encryption

import (
	"bytes"
	"testing"
)

func TestAesCfbEncipher(t *testing.T) {
	raw := []byte("是鸟也海运则将徙于南冥")
	cfb, _ := NewAesCfbEncipher("😉")
	cipher := cfb.Encrypt(raw)
	if bytes.Equal(raw, cipher) {
		t.FailNow()
	}
	plain := cfb.Decrypt(cipher)
	if !bytes.Equal(raw, plain) {
		t.FailNow()
	}

}
