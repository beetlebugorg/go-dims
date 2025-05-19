package main

import (
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"

	"github.com/beetlebugorg/go-dims/pkg/dims"
)

type EncryptionCmd struct {
	URL string `arg:"" help:"The URL to encrypt."`
}

func (e *EncryptionCmd) Run() error {
	if e.URL == "" {
		return errors.New("URL is required")
	}

	config := core.ReadConfig()

	result, err := dims.EncryptURL(config.SigningKey, e.URL)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	fmt.Println("Encrypted URL:", result)

	return nil
}

type DecryptionCmd struct {
	URL string `arg:"" help:"The URL to decrypt."`
}

func (e *DecryptionCmd) Run() error {
	if e.URL == "" {
		return errors.New("URL is required")
	}

	config := core.ReadConfig()

	result, err := dims.DecryptURL(config.SigningKey, e.URL)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	fmt.Println("Decrypted URL:", result)

	return nil
}
