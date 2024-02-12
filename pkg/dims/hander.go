package dims

import (
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func NewHandler() http.Handler {
	imagick.Initialize()

	config := ReadConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/dims4/{clientId}/{signature}/{timestamp}/{commands...}", func(w http.ResponseWriter, r *http.Request) {
		handleDims4(config, w, r)
	})

	return mux
}

func handleDims4(config Config, w http.ResponseWriter, r *http.Request) {
	slog.Info("handleDims5()",
		"url", r.URL,
		"clientId", r.PathValue("clientId"),
		"signature", r.PathValue("signature"),
		"timestamp", r.PathValue("timestamp"),
		"commands", r.PathValue("commands"))

	var timestamp int32
	fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)

	hash := md5.New()
	io.WriteString(hash, r.PathValue("clientId"))
	io.WriteString(hash, r.PathValue("commands"))
	io.WriteString(hash, r.URL.Query().Get("url"))

	request := Request{
		clientId:    r.PathValue("clientId"),
		imageUrl:    r.URL.Query().Get("url"),
		timestamp:   timestamp,
		noImageUrl:  config.NoImageUrl,
		commands:    r.PathValue("commands"),
		config:      config,
		requestHash: fmt.Sprintf("%x", hash.Sum(nil)),
		signature:   r.PathValue("signature"),
	}

	// Verify signature.
	if err := request.verifySignature(); err != nil {
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// Download image.
	if err := request.fetchImage(); err != nil {
		slog.Error("downloadImage failed.", "error", err)

		http.Error(w, "Failed to download image", request.status)
		return
	}

	// Execute Imagemagick commands.
	imageType, imageBlob, err := request.processImage()
	if err != nil {
		slog.Error("executeImagemagick failed.", "error", err)

		http.Error(w, "Failed to execute Imagemagick", http.StatusInternalServerError)
		return
	}

	// Serve the image.
	if err := request.sendImage(w, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}
