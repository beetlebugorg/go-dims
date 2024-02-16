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
	"crypto/md5"
	"fmt"
	"io"
	"strings"
)

type MD5Algorithm struct {
	key       string
	timestamp int32
}

// NewMD5 returns a new MD5Algorithm.
func NewMD5(signingKey string, timestamp int32) SignatureAlgorithm {
	return MD5Algorithm{
		key:       signingKey,
		timestamp: timestamp,
	}
}

// Sign returns a signed string using the MD5 algorithm.
func (h MD5Algorithm) Sign(commands []Command, imageUrl string) string {
	// Concatenate the commands into a single string.
	commandStrings := make([]string, len(commands))
	for _, command := range commands {
		commandStrings = append(commandStrings, fmt.Sprintf("%s/%s", command.Name, command.Args))
	}

	sanitizedCommands := strings.Join(commandStrings, "/")
	sanitizedCommands = strings.ReplaceAll(sanitizedCommands, " ", "+")
	sanitizedCommands = strings.Trim(sanitizedCommands, "/")

	timestamp := fmt.Sprintf("%d", h.timestamp)

	hash := md5.New()
	io.WriteString(hash, timestamp)
	io.WriteString(hash, h.key)
	io.WriteString(hash, sanitizedCommands)
	io.WriteString(hash, imageUrl)

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}
