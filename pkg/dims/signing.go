package dims

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/signing"
)

func SignUrl(imageUrl string) (string, error) {
	config := core.ReadConfig()

	if config.SigningKey == "" {
		return "", fmt.Errorf("signing key is required")
	}

	signer, err := signing.NewSigner(imageUrl, *config)
	if err != nil {
		return "", err
	}

	return signer.SignedUrl(), nil
}
