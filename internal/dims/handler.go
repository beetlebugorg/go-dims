package dims

import (
	"log/slog"
	"net/http"
)

func Handler(kernel Kernel, config Config, w http.ResponseWriter, r *http.Request) {
	slog.Info("handleDims5()",
		"imageUrl", r.URL.Query().Get("url"),
		"clientId", r.PathValue("clientId"),
		"signature", r.PathValue("signature"),
		"timestamp", r.PathValue("timestamp"),
		"commands", r.PathValue("commands"))

	// Verify signature.
	if !config.DevelopmentMode {
		if !kernel.ValidateSignature() {
			kernel.SendError(w, 500, "signature verification failed.")

			return
		}
	}

	// Download image.
	if err := kernel.FetchImage(); err != nil {
		slog.Error("downloadImage failed.", "error", err)

		kernel.SendError(w, 500, "downloadImage failed.")

		return
	}

	// Execute Imagemagick commands.
	imageType, imageBlob, err := kernel.ProcessImage()
	if err != nil {
		slog.Error("executeImagemagick failed.", "error", err)

		kernel.SendError(w, 500, "image processing failed.")

		return
	}

	// Serve the image.
	if err := kernel.SendImage(w, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}
