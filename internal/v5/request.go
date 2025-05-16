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
	"github.com/beetlebugorg/go-dims/internal/core"
	dims "github.com/beetlebugorg/go-dims/internal/http"
	"log/slog"
	"net/http"
)

type Request struct {
	*dims.Request
}

func NewRequest(r *http.Request, w http.ResponseWriter, config core.Config) (*Request, error) {
	request, err := dims.NewRequest(r, w, config)
	if err != nil {
		return nil, err
	}

	request.Signature = r.URL.Query().Get("sig")

	return &Request{
		Request: request,
	}, nil
}

// Validate verifies the signature of the image resize is valid.
func (v5 *Request) Validate() bool {
	return ValidateSignature(v5.Signature, v5.ImageUrl, v5.SignedParams, v5.RawCommands, v5.Config().SigningKey)
}

func ValidateSignature(signature, imageUrl string, signedParams map[string]string, command string, signingKey string) bool {
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(command))
	mac.Write([]byte(imageUrl))

	for _, signedParam := range signedParams {
		mac.Write([]byte(signedParam))
	}

	expectedSignature := mac.Sum(nil)[0:31]
	gotSignature, err := hex.DecodeString(signature)
	if err != nil {
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
