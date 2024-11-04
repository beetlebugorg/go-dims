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
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"github.com/charmbracelet/lipgloss"
	"log/slog"
	"net/url"
	"strings"
)

var keyColor = lipgloss.NewStyle().
	Bold(false).
	Foreground(lipgloss.Color("#517B59"))

var valueColor = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#6DB6FF"))

func SignUrl(requestUrl string, devMode bool) (string, error) {
	u, err := url.Parse(requestUrl)
	if err != nil {
		slog.Error("NewRequestFromUrl failed.", "error", err)
		return "", errors.New("failed to parse request URL")
	}

	// Determine if its a v5 or v5 request
	var signer dims.UrlSigner
	if strings.HasPrefix(u.Path, "/v5/dims") {
		signer, err = v5.NewSigner(u)
		if err != nil {
			return "", err
		}
	} else {
		signer, err = v4.NewSigner(u)
		if err != nil {
			return "", err
		}
	}

	if devMode {
		u.Scheme = "http"
		u.Host = "localhost:8080"
	}

	fmt.Println("Image to be transformed:")
	fmt.Printf("\n%s\n\n", valueColor.Render(signer.ImageUrl()))

	commands := signer.Commands()
	if commands != nil {
		fmt.Printf("Transformation commands found:\n\n")
		for _, command := range commands {
			fmt.Printf("  %s('%s')\n",
				valueColor.Render(command.Name),
				keyColor.Render(command.Args))
		}
	}

	return signer.SignUrl()
}
