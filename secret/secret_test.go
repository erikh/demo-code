package secret

import (
	"bytes"
	"testing"
)

func TestSecretCrypto(t *testing.T) {
	m, err := NewMessage([]byte("secret"), false).Encrypt([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(m.Value(), []byte("secret")) {
		t.Fatal("bytes were equal after encryption")
	}

	if !m.Encrypted() {
		t.Fatal("was not flagged as encrypted when should have been")
	}

	_, err = m.Decrypt([]byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"))
	if err == nil {
		t.Fatal("successfully decrypted with invalid key")
	}

	m2, err := m.Decrypt([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	if err != nil {
		t.Fatal("could not decrypt")
	}

	if m2.Encrypted() {
		t.Fatal("was flagged as encrypted when should not have been")
	}

	if !bytes.Equal(m2.Value(), []byte("secret")) {
		t.Fatal("bytes were not equal after decryption")
	}
}
