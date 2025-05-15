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
	"net/http"
	"time"
)

type HealthCmd struct {
}

func (h *HealthCmd) Run() error {
	config := core.ReadConfig()

	url := fmt.Sprintf("http://localhost%s/healthz", config.BindAddress)
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %v", err)
	}

	return nil
}
