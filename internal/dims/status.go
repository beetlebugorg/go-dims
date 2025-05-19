package dims

import (
	"github.com/beetlebugorg/go-dims/internal/core"
	"net/http"
)

func HandleDimsStatus(config core.Config, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ALIVE"))
}
