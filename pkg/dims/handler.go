package dims

import (
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/v4"
	"github.com/beetlebugorg/go-dims/internal/v5"
	"log/slog"
	"net/http"

	"github.com/beetlebugorg/go-dims/internal/dims"
	_ "github.com/beetlebugorg/go-dims/internal/source"
)

func NewHandler(config core.Config) http.Handler {

	mux := http.NewServeMux()

	slog.Debug("startup", "config", config)

	// v4 endpoint
	mux.HandleFunc("/dims4/{clientId}/{signature}/{timestamp}/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			config.EtagAlgorithm = "md5"

			request, err := v4.NewRequest(r, w, config)
			if err != nil {
				if err := request.SendError(err); err != nil {
					slog.Error("error sending error response", "error", err)
				}
				return
			}

			if err := dims.Handler(request); err != nil {
				if err := request.SendError(err); err != nil {
					slog.Error("error sending error response", "error", err)
				}
				return
			}
		})

	// v5 endpoint
	mux.HandleFunc("/v5/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			config.EtagAlgorithm = "hmac-sha256"

			request, err := v5.NewRequest(r, w, config)
			if err != nil {
				if err := request.SendError(err); err != nil {
					slog.Error("error sending error response", "error", err)
				}
				return
			}

			if err := dims.Handler(request); err != nil {
				if err := request.SendError(err); err != nil {
					slog.Error("error sending error response", "error", err)
				}
				return
			}
		})

	mux.HandleFunc("/dims-status/",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(config, w, r)
		})
	mux.HandleFunc("/healthz",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(config, w, r)
		})

	return mux
}
