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

package v4

import (
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/core"
	"github.com/beetlebugorg/go-dims/internal/dims/request"
)

func ParseAndValidateV4Request(r *http.Request, config core.Config) (*request.Request, error) {
	clientId := r.PathValue("clientId")
	timestamp := r.PathValue("timestamp")
	signature := r.PathValue("signature")
	imageUrl := r.URL.Query().Get("url")
	commands := r.PathValue("commands")

	// Validate signature
	if !config.DevelopmentMode &&
		!validateSignatureV4(commands, timestamp, imageUrl, config.SigningKey, signature) {
		return nil, &core.StatusError{
			StatusCode: http.StatusUnauthorized,
			Message:    "invalid signature",
		}
	}

	h := md5.New()
	h.Write([]byte(clientId))
	h.Write([]byte(commands))
	h.Write([]byte(imageUrl))
	id := fmt.Sprintf("%x", h.Sum(nil))

	return request.NewDimsRequest(*r, id, imageUrl, commands, config), nil
}

// ValidateSignature verifies the signature of the image resize is valid.
func validateSignatureV4(commands string, timestamp string, imageUrl string, signingKey string, signature string) bool {
	slog.Debug("verifySignature", "url", imageUrl)

	sanitizedCommands := strings.ReplaceAll(commands, " ", "+")
	sanitizedCommands = strings.Trim(sanitizedCommands, "/")
	sanitizedCommands += "/"

	hash := md5.New()
	hash.Write([]byte(timestamp))
	hash.Write([]byte(signingKey))
	hash.Write([]byte(sanitizedCommands))
	hash.Write([]byte(imageUrl))

	expectedSignature := fmt.Sprintf("%x", hash.Sum(nil))[0:7]

	if expectedSignature == signature {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", expectedSignature,
		"got", signature)

	return false
}
