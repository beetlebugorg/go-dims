package dims

import (
	"log/slog"
	"net/http"
)

func Handler(request Request, w http.ResponseWriter, r *http.Request) {
	slog.Info("handleDims5()",
		"imageUrl", r.URL.Query().Get("url"),
		"clientId", r.PathValue("clientId"),
		"signature", r.PathValue("signature"),
		"timestamp", r.PathValue("timestamp"),
		"commands", r.PathValue("commands"))

	// Verify signature.
	if !request.DevMode {
		if err := request.VerifySignature(); err != nil {
			request.SourceImage = Image{
				Status: 500,
			}
			request.SendError(w)

			return
		}
	}

	// Download image.
	if err := request.FetchImage(); err != nil {
		slog.Error("downloadImage failed.", "error", err)

		request.SendError(w)

		return
	}

	// Execute Imagemagick commands.
	imageType, imageBlob, err := request.ProcessImage()
	if err != nil {
		slog.Error("executeImagemagick failed.", "error", err)

		request.SendError(w)

		return
	}

	// Serve the image.
	if err := request.SendImage(w, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}
