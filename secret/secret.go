package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Message is a secret message. It may be in an encrypted or decrypted state,
// indicated by the Encrypted() flag.
type Message struct {
	value     []byte
	encrypted bool
}

// NewMessage creates a new message.
func NewMessage(value []byte, encrypted bool) Message {
	return Message{
		value:     value,
		encrypted: encrypted,
	}
}

// Value returns the value of the message.
func (m Message) Value() []byte {
	return m.value
}

// Encrypted returns whether or not the message is encrypted.
func (m Message) Encrypted() bool {
	return m.encrypted
}

// Encrypt encrypts the message.
func (m Message) Encrypt(key []byte) (Message, error) {
	m2 := Message{}

	c, err := aes.NewCipher(key)
	if err != nil {
		return m2, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return m2, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return m2, err
	}

	m2.value = gcm.Seal(nonce, nonce, m.value, nil)
	m2.encrypted = true
	return m2, nil
}

// Decrypt decrypts the message.
func (m Message) Decrypt(key []byte) (Message, error) {
	m2 := Message{}

	c, err := aes.NewCipher(key)
	if err != nil {
		return m2, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return m2, err
	}

	ns := gcm.NonceSize()
	if len(m.value) < ns {
		return m2, errors.New("ciphertext is too small")
	}

	nonce, text := m.value[:ns], m.value[ns:]

	m2.value, err = gcm.Open(nil, nonce, text, nil)
	if err != nil {
		return m2, err
	}
	m2.encrypted = false
	return m2, nil
}
