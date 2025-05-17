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
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/pkg/dims"
	"io"
	"net/url"
	"os"
	"strings"
)

type SignCmd struct {
	ImageURL     string `arg:"" name:"imageUrl" help:"Image URL to sign. For v4 urls place any value in the signature position in the URL."`
	KeyFile      string `help:"Path to the key file."`
	KeyFromStdin bool   `help:"Read the key from standard input."`
	Encrypt      bool   `help:"Encrypt the Image URL."`
}

func (cmd *SignCmd) Run() error {
	var signingKey string
	if cmd.KeyFile != "" {
		keyBytes, err := os.ReadFile(cmd.KeyFile)
		if err != nil {
			return err
		}

		signingKey = string(keyBytes)
	} else if cmd.KeyFromStdin {
		keyBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			os.Exit(1)
		}

		signingKey = string(keyBytes)
	}

	config := core.ReadConfig()
	config.SigningKey = strings.Trim(signingKey, "\n\r ")

	signedUrl, err := dims.SignUrl(cmd.ImageURL)
	if err != nil {
		if strings.Contains(err.Error(), "signing key is required") {
			fmt.Println("Signing key is required. Use --key-file or --key-from-stdin to provide a key. Or set the DIMS_SIGNING_KEY environment variable.")
			return nil
		}
		return err
	}

	if cmd.Encrypt {
		config := core.ReadConfig()

		u, err := url.Parse(signedUrl)
		if err != nil {
			return err
		}

		query := u.Query()
		imageUrl := query.Get("url")
		encryptedUrl, err := dims.EncryptURL(config.SigningKey, imageUrl)
		if err != nil {
			return err
		}
		query.Set("eurl", encryptedUrl)
		query.Del("url")
		u.RawQuery = query.Encode()

		signedUrl = u.String()
	}

	fmt.Printf("\n%s\n", signedUrl)

	return nil
}
