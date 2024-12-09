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

package v4

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims"
)

type RequestV4 struct {
	dims.Request
	Timestamp     int32
	TrailingSlash bool
}

func NewRequest(r *http.Request, config dims.Config) *RequestV4 {
	var timestamp int32
	n, err := fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)
	if err != nil || n != 1 {
		timestamp = 0
	}

	h := md5.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestV4{
		dims.Request{
			Id:          requestHash,
			Config:      config,
			ClientId:    r.PathValue("clientId"),
			ImageUrl:    r.URL.Query().Get("url"),
			RawCommands: r.PathValue("commands"),
			Signature:   r.PathValue("signature"),
		},
		timestamp,
		strings.HasSuffix(r.URL.Path, "/"),
	}
}

// ValidateSignature verifies the signature of the image resize is valid.
func (r *RequestV4) ValidateSignature() bool {
	slog.Debug("verifySignature", "url", r.ImageUrl)

	signature := r.Sign()

	if bytes.Equal([]byte(signature), []byte(r.Signature)) {
		return true
	}

	slog.Error("verifySignature failed.", "expected", signature, "got", r.Signature)

	return false
}

// Sign returns a signed string using the MD5 algorithm.
func (r *RequestV4) Sign() string {
	sanitizedCommands := strings.ReplaceAll(r.Request.RawCommands, " ", "+")
	sanitizedCommands = strings.Trim(sanitizedCommands, "/")

	// This makes the signing algorithm compatible with mod-dims.
	if r.TrailingSlash {
		sanitizedCommands += "/"
	}

	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d", r.Timestamp)))
	hash.Write([]byte(r.Config.Signing.SigningKey))
	hash.Write([]byte(sanitizedCommands))
	hash.Write([]byte(r.Request.ImageUrl))

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
