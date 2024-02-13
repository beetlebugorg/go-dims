package dims

import (
	"fmt"
	"net/http"

	"github.com/beetlebugorg/go-dims/internal/dims"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func NewHandler(debug bool, dev bool) http.Handler {
	imagick.Initialize()

	config := dims.ReadConfig()

	if debug {
		fmt.Printf("config: %+v\n", config)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/dims4/{clientId}/{signature}/{timestamp}/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDims4(config, debug, dev, w, r)
		})

	mux.HandleFunc("/dims-sizer/{url}",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsSizer(config, debug, dev, w, r)
		})

	mux.HandleFunc("/dims-status",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(config, debug, dev, w, r)
		})

	return mux
}
