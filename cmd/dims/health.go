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
