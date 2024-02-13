package dims

import (
	"github.com/beetlebugorg/go-dims/internal/dims"
)

func Sign(timestamp string, secret string, commands string, imageUrl string) string {
	return dims.Sign(timestamp, secret, commands, imageUrl)
}
