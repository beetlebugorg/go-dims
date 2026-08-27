package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v10"
	"io"
	"strings"
)

var encryptionKey []byte
var salt = []byte("go-dims")

func init() {
	envConfig := Signing{}
	RecordStartupError(env.Parse(&envConfig))

	var err error
	if encryptionKey, err = deriveKey(envConfig.SigningKey); err != nil {
		RecordStartupError(fmt.Errorf("failed to derive the encryption key: %w", err))
		return
	}
}

// LegacyKey reports whether the signing key uses the mod_dims sha1: prefix.
// That path derives the AES key from 16 hex characters, so it carries 64 bits
// of key material spread across 16 bytes.
func LegacyKey(secretKey string) bool {
	return strings.HasPrefix(secretKey, "sha1:")
}

func deriveKey(secretKey string) ([]byte, error) {
	if strings.HasPrefix(secretKey, "sha1:") {
		secret := secretKey[5:]
		hash := sha1.Sum([]byte(secret))          // returns [20]byte
		hexEncoded := hex.EncodeToString(hash[:]) // 40 hex chars
		keyFragment := strings.ToUpper(hexEncoded[:16])

		return []byte(keyFragment), nil
	} else {
		hash := sha256.New
		secret := secretKey
		if strings.HasPrefix(secret, "hkdf:") {
			secret = secret[5:]
		}
		return hkdf.Key(hash, []byte(secret), salt, "", 16)
	}
}

func EncryptURLKey(secretKey string, url string) (string, error) {
	key, err := deriveKey(secretKey)
	if err != nil {
		return "", err
	}

	return EncryptAES128GCM(key, url)
}

func DecryptURL(url string) (string, error) {
	url = strings.ReplaceAll(url, " ", "+")

	return DecryptAES128GCM(encryptionKey, url)
}

// DecryptURLKey decrypts the given eurl string using a derived AES-128-GCM key.
func DecryptURLKey(secretKey string, url string) (string, error) {
	key, err := deriveKey(secretKey)
	if err != nil {
		return "", err
	}

	url = strings.ReplaceAll(url, " ", "+")

	return DecryptAES128GCM(key, url)
}

// decodeEurl accepts the standard alphabet, which is what EncryptAES128GCM
// produces, and the URL safe alphabet, which a caller may reasonably send
// instead. Padding is optional in both.
func decodeEurl(value string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	for _, encoding := range encodings {
		if decoded, err := encoding.DecodeString(value); err == nil {
			return decoded, nil
		}
	}

	return nil, errors.New("eurl is not valid base64")
}

// DecryptAES128GCM takes a base64-encoded ciphertext and decrypts it using AES-128-GCM.
// The input must be encoded as: IV (12 bytes) | Ciphertext | Tag (16 bytes).
func DecryptAES128GCM(key []byte, base64EncryptedText string) (string, error) {
	encryptedData, err := decodeEurl(base64EncryptedText)
	if err != nil {
		return "", err
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(encryptedData) < nonceSize+aesgcm.Overhead() {
		return "", errors.New("invalid encrypted data length")
	}

	// Open expects the nonce, then the ciphertext with its tag still attached,
	// which is exactly the layout after the nonce.
	plaintext, err := aesgcm.Open(nil, encryptedData[:nonceSize], encryptedData[nonceSize:], nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptAES128GCM encrypts the given plaintext using AES-128-GCM with the provided key.
// The result is base64-encoded and includes IV (12 bytes) | ciphertext | tag (16 bytes).
func EncryptAES128GCM(key []byte, plaintext string) (string, error) {
	if len(key) != 16 {
		return "", errors.New("key must be 16 bytes for AES-128")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a 12-byte IV (nonce)
	iv := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := aesgcm.Seal(nil, iv, []byte(plaintext), nil) // ciphertext includes the tag at the end

	// Prepend IV to ciphertext+tag
	result := append(iv, ciphertext...)

	// Base64 encode the result
	encoded := base64.StdEncoding.EncodeToString(result)

	return encoded, nil
}
