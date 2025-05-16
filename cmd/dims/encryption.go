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
