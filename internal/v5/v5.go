// Copyright 2024 Jeremy Collins. All rights reserved.
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

package v5

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	core2 "github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/request"
	"log/slog"
	"net/http"
	"strings"
)

func ParseAndValidateV5Request(r *http.Request, w http.ResponseWriter, config core2.Config) (*request.HttpDimsRequest, error) {
	signature := r.PathValue("signature")
	imageUrl := r.URL.Query().Get("url")
	commands := r.PathValue("commands")

	eurl := r.URL.Query().Get("eurl")
	if eurl != "" {
		decryptedUrl, err := core2.DecryptURL(config.SigningKey, eurl)
		if err != nil {
			slog.Error("DecryptURL failed.", "error", err)
			return nil, fmt.Errorf("DecryptURL failed: %w", err)
		}

		imageUrl = decryptedUrl
	}

	h := sha256.New()
	h.Write([]byte(commands))
	h.Write([]byte(imageUrl))
	id := fmt.Sprintf("%x", h.Sum(nil))

	// Signed Parameters
	// _keys query parameter is a comma delimted list of keys to include in the signature.
	var signedKeys []string
	params := r.URL.Query().Get("_keys")
	if params != "" {
		keys := strings.Split(params, ",")
		for _, key := range keys {
			value := r.URL.Query().Get(key)
			if value != "" {
				signedKeys = append(signedKeys, value)
			}
		}
	}

	// Validate signature
	if !config.DevelopmentMode &&
		!ValidateSignatureV5(commands, imageUrl, signedKeys, config.SigningKey, signature) {
		return nil, &core2.StatusError{
			StatusCode: http.StatusUnauthorized,
			Message:    "invalid signature",
		}
	}

	return request.NewHttpDimsRequest(*r, w, id, imageUrl, commands, config), nil
}

// ValidateSignature verifies the signature of the image resize is valid.
func ValidateSignatureV5(commands string, imageUrl string, signedParams []string, signingKey string, signature string) bool {
	slog.Debug("verifySignature", "url", imageUrl)

	sanitizedArgs := strings.ReplaceAll(commands, " ", "+")

	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(sanitizedArgs))
	mac.Write([]byte(imageUrl))

	// _keys query parameter is a comma delimted list of keys to include in the signature.
	for _, signedParam := range signedParams {
		mac.Write([]byte(signedParam))
	}

	expectedSignature := mac.Sum(nil)[0:31]
	gotSignature, err := hex.DecodeString(signature)
	if err != nil {
		slog.Error("verifySignature failed.", "error", err)
		return false
	}

	if hmac.Equal(expectedSignature, gotSignature) {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", hex.EncodeToString(expectedSignature),
		"got", signature)

	return false
}
