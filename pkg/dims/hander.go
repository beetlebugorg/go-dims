package dims

import (
	"fmt"
	"net/http"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func NewHandler(debug bool, dev bool) http.Handler {
	imagick.Initialize()

	config := ReadConfig()

	if debug {
		fmt.Printf("config: %+v\n", config)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/dims4/{clientId}/{signature}/{timestamp}/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			handleDims4(config, debug, dev, w, r)
		})

	mux.HandleFunc("/dims-sizer/{url}",
		func(w http.ResponseWriter, r *http.Request) {
			handleDimsSizer(config, debug, dev, w, r)
		})

	mux.HandleFunc("/dims-status",
		func(w http.ResponseWriter, r *http.Request) {
			handleDimsStatus(config, debug, dev, w, r)
		})

	return mux
}
