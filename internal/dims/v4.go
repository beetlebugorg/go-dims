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

package dims

import (
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/core"
)

func ParseAndValidateV4Request(r *http.Request, config core.Config) (*Request, error) {
	h := md5.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	timestamp, err := strconv.ParseInt(r.PathValue("timestamp"), 10, 64)
	if err != nil {
		slog.Error("ParseInt failed.", "error", err)
		return nil, fmt.Errorf("ParseInt failed: %w", err)
	}

	request := Request{
		Id:          requestHash,
		Config:      config,
		ClientId:    r.PathValue("clientId"),
		ImageUrl:    r.URL.Query().Get("url"),
		RawCommands: r.PathValue("commands"),
		Timestamp:   timestamp,
		Signature:   r.PathValue("signature"),
	}

	// Validate signature
	if !config.DevelopmentMode && !validateSignatureV4(request) {
		return &request, &core.StatusError{
			StatusCode: http.StatusUnauthorized,
			Message:    "invalid signature",
		}
	}

	return &request, nil
}

// ValidateSignature verifies the signature of the image resize is valid.
func validateSignatureV4(request Request) bool {
	slog.Debug("verifySignature", "url", request.ImageUrl)

	expectedSignature := Sign_md5(request)

	if expectedSignature == request.Signature {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", expectedSignature,
		"got", request.Signature)

	return false
}

// Sign returns a signed string using the MD5 algorithm.
func Sign_md5(r Request) string {
	sanitizedCommands := strings.ReplaceAll(r.RawCommands, " ", "+")
	sanitizedCommands = strings.Trim(sanitizedCommands, "/")
	sanitizedCommands += "/"

	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d", r.Timestamp)))
	hash.Write([]byte(r.Config.Signing.SigningKey))
	hash.Write([]byte(sanitizedCommands))
	hash.Write([]byte(r.ImageUrl))

	slog.Debug("Sign_md5", "timestamp", r.Timestamp,
		"signingKey", r.Config.Signing.SigningKey,
		"sanitizedCommands", sanitizedCommands,
		"imageUrl", r.ImageUrl)

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}
