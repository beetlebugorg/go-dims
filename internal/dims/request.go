package dims

import (
	"context"
	"errors"
	"github.com/beetlebugorg/go-dims/internal/commands"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"net/url"
	"runtime/trace"
	"strings"
	"time"
)

type Request struct {
	URL                    *url.URL          // The URL of the http.
	ImageUrl               string            // The image URL that is being manipulated.
	SendContentDisposition bool              // The content disposition of the http.
	RawCommands            string            // The commands ('resize/100x100', 'strip/true/format/png', etc).
	Signature              string            // The signature of the request.
	SignedParams           map[string]string // The query parameters used to sign the request.
	SourceImage            core.Image        // The source image.
	config                 core.Config       // The global configuration.
	shrinkFactor           int
}

func NewRequest(url *url.URL, cmds string, config core.Config) (*Request, error) {
	imageUrl := url.Query().Get("url")
	eurl := url.Query().Get("eurl")
	if eurl != "" {
		decryptedUrl, err := core.DecryptURL(eurl)
		if err != nil {
			slog.Error("failed to decrypt eurl, ensure DIMS_SIGNING_KEY matches key used to encrypt. For mod_dims compatibility you must prepend 'sha1:' to the key.", "error", err)
			return &Request{}, err
		}

		imageUrl = decryptedUrl
	}

	// Signed Parameters
	// Include all parameters except for the signature, the image URL, and "eurl".
	var signedParams = make(map[string]string)
	for key, _ := range url.Query() {
		value := url.Query().Get(key)
		if key != "sig" && key != "eurl" && key != "_keys" && key != "url" && key != "download" {
			if value != "" {
				signedParams[key] = value
			}
		}
	}

	var sendContentDisposition = config.IncludeDisposition
	if url.Query().Get("download") == "1" || url.Query().Get("download") == "true" {
		sendContentDisposition = true
	}

	return &Request{
		URL:                    url,
		ImageUrl:               imageUrl,
		RawCommands:            cmds,
		SignedParams:           signedParams,
		SendContentDisposition: sendContentDisposition,
		config:                 config,
	}, nil
}

func (r *Request) Config() core.Config {
	return r.config
}

func (r *Request) LoadImage(sourceImage *core.Image) (*vips.ImageRef, error) {
	image, err := vips.NewImageFromBuffer(sourceImage.Bytes)
	if err != nil {
		return nil, err
	}
	importParams := vips.NewImportParams()
	importParams.AutoRotate.Set(true)

	r.shrinkFactor = 1
	requestedSize, err := r.requestedImageSize()
	if err == nil && vips.DetermineImageType(sourceImage.Bytes) == vips.ImageTypeJPEG {
		xs := image.Width() / int(requestedSize.Width)
		ys := image.Height() / int(requestedSize.Height)

		if (xs > 2) || (ys > 2) {
			importParams.JpegShrinkFactor.Set(4)
			r.shrinkFactor = 4
		}
	}

	return vips.LoadImageFromBuffer(sourceImage.Bytes, importParams)
}

// ProcessImage will execute the commands on the image.
func (r *Request) ProcessImage(image *vips.ImageRef, errorImage bool) (string, []byte, error) {
	ctx := context.Background()

	// Execute the commands.
	ctx, task := trace.NewTask(ctx, "v5.ProcessImage")
	defer task.End()

	opts := commands.ExportOptions{
		ImageType:        r.outputFormat(),
		JpegExportParams: core.NewJpegExportParams(r.config.ImageOutputOptions.Jpeg, r.config.StripMetadata),
		PngExportParams:  core.NewPngExportParams(r.config.ImageOutputOptions.Png, r.config.StripMetadata),
		WebpExportParams: core.NewWebpExportParams(r.config.ImageOutputOptions.Webp, r.config.StripMetadata),
		GifExportParams:  vips.NewGifExportParams(),
		TiffExportParams: vips.NewTiffExportParams(),
	}

	stripMetadata := r.config.StripMetadata
	opts.GifExportParams.StripMetadata = stripMetadata
	opts.TiffExportParams.StripMetadata = stripMetadata

	for _, command := range r.Commands() {
		region := trace.StartRegion(ctx, command.Name)

		if operation, ok := commands.VipsTransformCommands[command.Name]; ok {
			if command.Name == "strip" && command.Args != "true" {
				stripMetadata = false
			}

			if err := operation(image, command.Args); err != nil && !errorImage {
				return "", nil, err
			}
		} else if operation, ok := commands.VipsExportCommands[command.Name]; ok {
			if err := operation(image, command.Args, &opts); err != nil && !errorImage {
				return "", nil, err
			}
		} else if operation, ok := commands.VipsRequestCommands[command.Name]; ok && !errorImage {
			if err := operation(image, command.Args, commands.RequestOperation{
				Config: r.config,
				URL:    r.URL,
			}); err != nil {
				return "", nil, err
			}
		}

		region.End()
	}

	if stripMetadata {
		if err := image.RemoveMetadata(); err != nil {
			return "", nil, err
		}
	}

	switch opts.ImageType {
	case vips.ImageTypeJPEG:
		imageBytes, _, err := image.ExportJpeg(opts.JpegExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeJPEG], imageBytes, nil

	case vips.ImageTypePNG:
		imageBytes, _, err := image.ExportPng(opts.PngExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypePNG], imageBytes, nil

	case vips.ImageTypeWEBP:
		imageBytes, _, err := image.ExportWebp(opts.WebpExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeWEBP], imageBytes, nil
	case vips.ImageTypeGIF:
		imageBytes, _, err := image.ExportGIF(opts.GifExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeGIF], imageBytes, nil
	case vips.ImageTypeTIFF:
		imageBytes, _, err := image.ExportTiff(opts.TiffExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeTIFF], imageBytes, nil
	}

	imageBytes, _, err := image.ExportNative()
	if err != nil {
		return "", nil, err
	}

	return vips.ImageTypes[opts.ImageType], imageBytes, nil
}

func (r *Request) FetchImage(timeout time.Duration) (*core.Image, error) {
	image, err := core.FetchImage(r.ImageUrl, timeout)
	if err != nil {
		return nil, err
	}

	r.SourceImage = *image

	return image, nil
}

func (r *Request) Commands() []commands.Command {
	cmds := make([]commands.Command, 0)
	parsedCommands := strings.Split(strings.Trim(r.RawCommands, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		cmds = append(cmds, commands.Command{
			Name: command,
			Args: args,
		})
	}

	return cmds
}

// Parse through the requested commands and return requested image size for thumbnail and resize
// commands.
//
// This is used while reading an image to improve performance when generating thumbnails from very
// large images.
func (r *Request) requestedImageSize() (geometry.Geometry, error) {
	for _, command := range r.Commands() {
		if command.Name == "thumbnail" || command.Name == "resize" {
			rect, err := geometry.ParseGeometry(command.Args)
			if err != nil {
				return geometry.Geometry{}, err
			}

			if rect.Width > 0 && rect.Height > 0 {
				return rect, nil
			}

		}
	}

	return geometry.Geometry{}, errors.New("no resize or thumbnail command found")
}

func (r *Request) outputFormat() vips.ImageType {
	// If default is configured, use that first.
	if r.config.OutputFormat.Default != "" {
		return core.ImageTypes[r.config.OutputFormat.Default]
	}

	// If not configured and image is either GIF or SVG, default to PNG, otherwise return "".
	if r.SourceImage.Format == vips.ImageTypeGIF || r.SourceImage.Format == vips.ImageTypeSVG {
		return vips.ImageTypePNG
	}

	return r.SourceImage.Format
}
