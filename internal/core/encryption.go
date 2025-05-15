// Copyright 2025 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"strings"
)

func EncryptionKey(secretKey string) []byte {
	// Step 1: SHA-1 hash the secret key
	hash := sha1.Sum([]byte(secretKey)) // returns [20]byte

	// Step 2: Convert the hash to hex
	hexEncoded := hex.EncodeToString(hash[:]) // 40 hex chars

	// Step 3: Use first 16 characters of the hex string as the key
	keyFragment := strings.ToUpper(hexEncoded[:16])
	return []byte(keyFragment)
}

func EncryptURL(secretKey string, url string) (string, error) {
	key := EncryptionKey(secretKey)

	encryptedURL, err := EncryptAES128GCM(key, url)
	if err != nil {
		return "", err
	}

	return encryptedURL, nil
}

// DecryptURL decrypts the given eurl string using a derived AES-128-GCM key.
func DecryptURL(secretKey string, base64Eurl string) (string, error) {
	key := EncryptionKey(secretKey)

	// Handle spaces in base64Eurl
	base64Eurl = strings.ReplaceAll(base64Eurl, " ", "+")

	return DecryptAES128GCM(key, base64Eurl)
}

// DecryptAES128GCM takes a base64-encoded ciphertext and decrypts it using AES-128-GCM.
// The input must be encoded as: IV (12 bytes) | Ciphertext | Tag (16 bytes).
func DecryptAES128GCM(key []byte, base64EncryptedText string) (string, error) {
	// Decode the base64 input
	encryptedData, err := base64.StdEncoding.DecodeString(base64EncryptedText)
	if err != nil {
		slog.Error("base64 decode failed.", "data", base64EncryptedText)
		return "", err
	}

	if len(encryptedData) < 12+16 {
		return "", errors.New("invalid encrypted data length")
	}

	// Extract IV, ciphertext, and tag
	iv := encryptedData[:12]
	tag := encryptedData[len(encryptedData)-16:]
	ciphertext := encryptedData[12 : len(encryptedData)-16]

	// Concatenate ciphertext and tag for Go's AEAD interface
	ciphertextWithTag := append(ciphertext, tag...)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, iv, ciphertextWithTag, nil)
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

	// Generate a 12-byte IV (nonce)
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
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

	// Encrypt
	ciphertext := aesgcm.Seal(nil, iv, []byte(plaintext), nil) // ciphertext includes the tag at the end

	// Prepend IV to ciphertext+tag
	result := append(iv, ciphertext...)

	// Base64 encode the result
	encoded := base64.StdEncoding.EncodeToString(result)

	return encoded, nil
}
