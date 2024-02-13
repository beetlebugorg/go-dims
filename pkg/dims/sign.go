package dims

import (
	"github.com/beetlebugorg/go-dims/internal/dims"
)

func Sign(timestamp string, secret string, commands string, imageUrl string) string {
	return dims.Sign(timestamp, secret, commands, imageUrl)
}

func SignHmacSha256(timestamp string, secret string, commands string, imageUrl string) string {
	return dims.SignHmacSha256(timestamp, secret, commands, imageUrl)
}
