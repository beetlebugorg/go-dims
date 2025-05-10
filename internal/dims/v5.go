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

package dims

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/core"
)

func ParseAndValidV5Request(r *http.Request, config core.Config) (*Request, error) {
	h := sha256.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	request := Request{
		HttpRequest: r,
		Id:          requestHash,
		Config:      config,
		ClientId:    r.URL.Query().Get("clientId"),
		ImageUrl:    r.URL.Query().Get("url"),
		RawCommands: r.PathValue("commands"),
		Signature:   r.URL.Query().Get("sig"),
	}

	// Validate signature
	if !config.DevelopmentMode && !ValidateSignature(request) {
		return nil, fmt.Errorf("signature mismatch")
	}

	return &request, nil
}

// ValidateSignature verifies the signature of the image resize is valid.
func ValidateSignature(request Request) bool {
	slog.Debug("verifySignature", "url", request.ImageUrl)

	expectedSignature := Sign_HmacSha256_128(request)
	gotSignature, err := hex.DecodeString(request.Signature)
	if err != nil {
		slog.Error("verifySignature failed.", "error", err)
		return false
	}

	if hmac.Equal(expectedSignature, gotSignature) {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", hex.EncodeToString(expectedSignature),
		"got", request.Signature)

	return false
}

// Sign returns a signed string using HMAC-SHA256-128.
func Sign_HmacSha256_128(request Request) []byte {
	sanitizedArgs := strings.ReplaceAll(request.RawCommands, " ", "+")

	mac := hmac.New(sha256.New, []byte(request.Config.SigningKey))
	mac.Write([]byte(sanitizedArgs))
	mac.Write([]byte(request.ImageUrl))

	return mac.Sum(nil)[0:31]
}
