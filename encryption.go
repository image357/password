package password

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"unicode/utf8"
)

const saltLength = 32

// HashFunc is a function signature.
// The Hash function will be called for password and secret hashing.
type HashFunc func(data []byte, salt []byte) [32]byte

// Hash will calculate a 32 byte hash from a given byte slice.
// It is used for password and secret hashing.
// You can overwrite it with any function that meets the HashFunc signature.
// By default, it is set to a variant of argon2.Key.
var Hash HashFunc = argon2iHash

func sha256Hash(data []byte, salt []byte) [32]byte {
	temp := make([]byte, 0, 8*saltLength)
	temp = append(temp, salt...)
	temp = append(temp, data...)
	return sha256.Sum256(temp)
}

const argon2iMemory uint32 = 32
const argon2iTime uint32 = 3
const argon2iThreads uint8 = 4

func argon2iHash(data []byte, salt []byte) [32]byte {
	return [32]byte(argon2.Key(data, salt, argon2iTime, argon2iMemory*1024, argon2iThreads, 32))
}

// getHashedPassword returns random salt + hash with base64 encoding.
func getHashedPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := Hash([]byte(password), salt)
	hashedPassword := append(salt, hash[:]...)

	return base64.StdEncoding.EncodeToString(hashedPassword), nil
}

// compareHashedPassword compares a hashed password A against a password B, by performing the hash operation on B with the same salt from A.
// The input hashedPassword must be base64 encoded.
func compareHashedPassword(hashedPassword string, password string) (bool, error) {
	// decode hashed password
	data1, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false, err
	}

	// extract salt
	if len(data1) < saltLength {
		return false, fmt.Errorf("hashed password is too short")
	}
	salt := make([]byte, saltLength)
	copy(salt, data1[:saltLength])

	// hash other password
	hash := Hash([]byte(password), salt)
	data2 := append(salt, hash[:]...)

	// compare
	result := subtle.ConstantTimeCompare(data1, data2) == 1

	return result, nil
}

// comparePassword compares two passwords with constant time
func comparePassword(password1 string, password2 string) bool {
	hash1 := sha256.Sum256([]byte(password1))
	hash2 := sha256.Sum256([]byte(password2))

	return subtle.ConstantTimeCompare(hash1[:], hash2[:]) == 1
}

// encrypt a given text with AES256 and return a base64 representation.
// The secret is hashed with the custom Hash function.
// Galois Counter Mode is used.
// The nonce is stored as a prefix of the ciphertext.
func encrypt(text string, secret string) (string, error) {
	// create salt
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// hash secret
	secretHash := Hash([]byte(secret), salt)

	// prepare cipher
	block, err := aes.NewCipher(secretHash[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// create nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	// encrypt
	encrypted := gcm.Seal(nil, nonce, []byte(text), nil)
	saltAndNonce := append(salt, nonce...)
	cipherBytes := append(saltAndNonce, encrypted...)

	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// decrypt a given ciphertext in base64 representation with AES256.
// The secret is hashed with the custom Hash function.
// Galois Counter Mode is used.
// The nonce is retrieved as a prefix of the ciphertext.
func decrypt(ciphertext string, secret string) (string, error) {
	// extract salt
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(cipherBytes) < saltLength {
		return "", fmt.Errorf("ciphertext is too short")
	}
	salt := cipherBytes[:saltLength]
	cipherBytes = cipherBytes[saltLength:]

	// hash secret
	secretHash := Hash([]byte(secret), salt)

	// prepare cipher
	block, err := aes.NewCipher(secretHash[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// extract nonce
	if len(cipherBytes) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext is too short")
	}
	nonce := cipherBytes[:gcm.NonceSize()]
	msg := cipherBytes[gcm.NonceSize():]

	// decrypt
	textBytes, err := gcm.Open(nil, nonce, msg, nil)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(textBytes) {
		return "", fmt.Errorf("invalid utf8 character after decrypt")
	}
	text := string(textBytes)

	return text, nil
}
