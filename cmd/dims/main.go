package main

import (
	"log/slog"
	"net/http"

	"github.com/beetlebugorg/go-dims/pkg/dims"
)

func main() {
	err := http.ListenAndServe(":8080", dims.NewHandler())
	if err != nil {
		slog.Error("Server failed.", "error", err)
		return
	}

	slog.Info("Server started.", "port", 8080)
}
