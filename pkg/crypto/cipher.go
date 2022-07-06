package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
)

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// use the hash of the input text to stabilize encryption, but still
	// don't leak any info (except file changes)
	hash := sha256.Sum256(text)
	nonce := hash[:gcm.NonceSize()]
	return gcm.Seal(nonce, nonce, text, nil), nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(text) < gcm.NonceSize() {
		return nil, errors.New("malformed text")
	}

	return gcm.Open(nil,
		text[:gcm.NonceSize()],
		text[gcm.NonceSize():],
		nil,
	)
}
