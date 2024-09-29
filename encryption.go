package password

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// encrypt a given text with AES256 and return a base64 representation.
// The secret is hashed with sha256.
// Galois Counter Mode is used.
// The nonce is stored as a prefix of the ciphertext.
func encrypt(text string, secret string) (string, error) {
	secretHash := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(secretHash[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}
	encrypted := gcm.Seal(nil, nonce, []byte(text), nil)
	cipherBytes := append(nonce, encrypted...)

	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// decrypt a given ciphertext in base64 representation with AES256.
// The secret is hashed with sha256.
// Galois Counter Mode is used.
// The nonce is retrieved as a prefix of the ciphertext.
func decrypt(ciphertext string, secret string) (string, error) {
	secretHash := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(secretHash[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if len(cipherBytes) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := cipherBytes[:gcm.NonceSize()]
	msg := cipherBytes[gcm.NonceSize():]
	text, err := gcm.Open(nil, nonce, msg, nil)
	if err != nil {
		return "", err
	}

	return string(text), nil
}
