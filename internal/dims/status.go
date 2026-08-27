package dims

import (
	"log/slog"
	"net/http"

	"github.com/beetlebugorg/go-dims/internal/core"
)

func HandleDimsStatus(config core.Config, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	if _, err := w.Write([]byte("ALIVE")); err != nil {
		slog.Debug("failed to write status response", "error", err)
	}
}
