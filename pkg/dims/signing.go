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
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"github.com/charmbracelet/lipgloss"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func SignUrl(requestUrl string, devMode bool) (string, error) {
	u, err := url.Parse(requestUrl)
	if err != nil {
		slog.Error("NewRequestFromUrl failed.", "error", err)
		return "", errors.New("failed to parse request URL")
	}

	// Determine if its a v5 or v5 request
	var resignedUrl *url.URL
	if strings.HasPrefix(u.Path, "/v5/dims") {
		resignedUrl, err = signV5(u)
		if err != nil {
			return "", err
		}
	} else {
		resignedUrl, err = signV4(u)
		if err != nil {
			return "", err
		}
	}

	if devMode {
		u.Scheme = "http"
		u.Host = "localhost:8080"
	}

	return resignedUrl.String(), nil
}

func signV4(u *url.URL) (*url.URL, error) {
	slog.Debug("signV4", "url", u.String(), "path", u.Path)

	path := strings.Replace(u.Path, "/v4/dims/", "", 1)
	path = strings.Replace(path, "/dims4/", "", 1)

	pattern, _ := regexp.Compile(`([^/]+)/(\w{7})/([^/]+)/(.+)`)
	match := pattern.FindStringSubmatch(path)
	if len(match) != 5 {
		slog.Error("signV4", "error", "failed to match path",
			"match", match, "path", u.Path)
		return nil, errors.New("failed to match path")
	}

	httpRequest := http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	httpRequest.SetPathValue("clientId", match[1])
	httpRequest.SetPathValue("signature", match[2])
	httpRequest.SetPathValue("timestamp", match[3])
	httpRequest.SetPathValue("commands", match[4])

	request := v4.NewRequest(&httpRequest, dims.Config{
		EnvironmentConfig: dims.ReadConfig(),
		DevelopmentMode:   false,
		DebugMode:         false,
	})

	var keyColor = lipgloss.NewStyle().
		Bold(false).
		Foreground(lipgloss.Color("#517B59"))

	var valueColor = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6DB6FF"))

	fmt.Println("Image to be transformed:")
	fmt.Printf("\n%s\n\n", valueColor.Render(request.ImageUrl))

	fmt.Printf("Transformation commands found:\n\n")
	for _, command := range request.Commands {
		fmt.Printf("%s('%s')\n",
			valueColor.Render(command.Name),
			keyColor.Render(command.Args))
	}

	signature := request.Sign()

	// Rebuild the URL with the new signature
	u.Path = fmt.Sprintf("/v4/dims/%s/%s/%s/%s", match[1], signature, match[3], match[4])

	return u, nil
}

func signV5(u *url.URL) (*url.URL, error) {
	httpRequest := http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	commands := strings.Replace(u.Path, "/v5/dims/", "", 1)
	commands = strings.Trim(commands, "/")
	httpRequest.SetPathValue("commands", commands)

	request := v5.NewRequest(&httpRequest, dims.Config{
		EnvironmentConfig: dims.ReadConfig(),
		DevelopmentMode:   false,
		DebugMode:         false,
	})

	var keyColor = lipgloss.NewStyle().
		Bold(false).
		Foreground(lipgloss.Color("#517B59"))

	var valueColor = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6DB6FF"))

	fmt.Println("Image to be transformed:")
	fmt.Printf("\n%s\n\n", valueColor.Render(request.ImageUrl))

	fmt.Printf("Transformation commands found:\n\n")
	for _, command := range request.Commands {
		fmt.Printf("%s('%s')\n",
			valueColor.Render(command.Name),
			keyColor.Render(command.Args))
	}

	signature := hex.EncodeToString(request.Sign())

	u.RawQuery = fmt.Sprintf("%s&sig=%s", u.RawQuery, signature)

	return u, nil
}
