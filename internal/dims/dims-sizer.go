package dims

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type sizerRequest struct {
	imageUrl    string
	sourceImage sizerSourceImage
}

type sizerSourceImage struct {
	image     []byte
	imageSize int
}

type sizerResponse struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func HandleDimsSizer(config Config, debug bool, dev bool, w http.ResponseWriter, r *http.Request) {
	request := sizerRequest{
		imageUrl: r.PathValue("url"),
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	request.fetchImage(request.imageUrl)

	// Read the image.
	mw.ReadImageBlob(request.sourceImage.image)

	w.Header().Set("Content-Type", "application/json")

	json, err := json.Marshal(sizerResponse{
		Height: int(mw.GetImageHeight()),
		Width:  int(mw.GetImageWidth()),
	})

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error marshalling JSON"))
	} else {
		w.WriteHeader(200)
		w.Write(json)
	}
}

func (r *sizerRequest) fetchImage(url string) error {
	image, err := http.Get(url)
	if err != nil || image.StatusCode != 200 {
		return errors.New("Failed to fetch image")
	}

	imageSize := int(image.ContentLength)
	imageBytes, _ := io.ReadAll(image.Body)

	r.sourceImage = sizerSourceImage{
		image:     imageBytes,
		imageSize: imageSize,
	}

	return nil
}
