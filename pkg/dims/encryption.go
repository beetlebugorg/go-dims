package dims

import (
	"github.com/beetlebugorg/go-dims/internal/core"
)

func EncryptURL(secretKey string, u string) (string, error) {
	return core.EncryptURLKey(secretKey, u)
}

// DecryptURL decrypts the given eurl string using a derived AES-128-GCM key.
func DecryptURL(secretKey string, base64Eurl string) (string, error) {
	return core.DecryptURLKey(secretKey, base64Eurl)
}
