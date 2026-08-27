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
	"slices"
	"strings"
	"time"
)

type Request struct {
	URL                    *url.URL    // The URL of the http.
	ImageUrl               string      // The image URL that is being manipulated.
	SendContentDisposition bool        // Whether to send a Content-Disposition header.
	Download               bool        // Whether the caller asked for a download rather than inline.
	RawCommands            string      // The commands ('resize/100x100', 'strip/true/format/png', etc).
	Signature              string      // The signature of the request.
	SignedParams           []string    // Values of every signed query parameter, ordered by name.
	LegacyParams           []string    // Values named by _keys, in _keys order. Legacy mode only.
	SourceImage            core.Image  // The source image.
	config                 core.Config // The global configuration.
	shrinkFactor           int
	commands               []commands.Command
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

	signedParams := SignedValues(url)
	legacyParams := LegacyValues(url)

	download := url.Query().Get("download") == "1" || url.Query().Get("download") == "true"

	// The header is sent when the deployment asks for it, or when this
	// request asks to download. Only the second makes it an attachment.
	sendContentDisposition := config.IncludeDisposition || download

	return &Request{
		URL:                    url,
		ImageUrl:               imageUrl,
		RawCommands:            cmds,
		SignedParams:           signedParams,
		LegacyParams:           legacyParams,
		SendContentDisposition: sendContentDisposition,
		Download:               download,
		config:                 config,
	}, nil
}

func (r *Request) Config() core.Config {
	return r.config
}

func (r *Request) LoadImage(sourceImage *core.Image) (*vips.ImageRef, error) {
	importParams := vips.NewImportParams()
	importParams.AutoRotate.Set(true)

	r.shrinkFactor = 1

	// Shrink on load helps only a JPEG that is much larger than the request.
	// Read the header for that case alone, so every other image loads once.
	requestedSize, err := r.requestedImageSize()
	if err == nil && vips.DetermineImageType(sourceImage.Bytes) == vips.ImageTypeJPEG {
		header, err := vips.NewImageFromBuffer(sourceImage.Bytes)
		if err != nil {
			return nil, err
		}

		xs := header.Width() / int(requestedSize.Width)
		ys := header.Height() / int(requestedSize.Height)
		header.Close()

		// Take the smaller ratio, so neither axis ends up below the requested
		// size and has to be scaled back up.
		if shrink := shrinkFactor(min(xs, ys)); shrink > 1 {
			importParams.JpegShrinkFactor.Set(shrink)
			r.shrinkFactor = shrink
		}
	}

	return vips.LoadImageFromBuffer(sourceImage.Bytes, importParams)
}

// shrinkFactor returns the largest power of two the JPEG loader can shrink by
// without dropping below the requested size. libvips supports 1, 2, 4, and 8.
func shrinkFactor(ratio int) int {
	switch {
	case ratio >= 8:
		return 8
	case ratio >= 4:
		return 4
	case ratio >= 2:
		return 2
	}

	return 1
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

		// An unknown name used to fall through all three maps and be ignored,
		// so a typo silently returned a differently processed image. The error
		// image is rendered with the same command list, so it is exempt.
		if !commands.Known(command.Name) && !errorImage {
			region.End()
			return "", nil, commands.NewOperationError(command.Name, command.Args, "unknown command")
		}

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

// Commands returns the parsed command list. The result is kept, since this is
// called once while loading the image and again while processing it.
func (r *Request) Commands() []commands.Command {
	if r.commands != nil {
		return r.commands
	}

	cmds := make([]commands.Command, 0)
	parsedCommands := strings.Split(strings.Trim(r.RawCommands, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		cmds = append(cmds, commands.Command{
			Name: parsedCommands[i],
			Args: parsedCommands[i+1],
		})
	}

	r.commands = cmds

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

// excludedFromSignature lists the query parameters that never take part in a
// signature. Every other parameter does.
var excludedFromSignature = map[string]bool{
	"sig":      true,
	"url":      true,
	"eurl":     true,
	"_keys":    true,
	"download": true,
}

// SignedValues returns the values of every query parameter that takes part in
// the signature, ordered by parameter name. A parameter carrying several
// values contributes each of them, in the order the URL gives them.
func SignedValues(u *url.URL) []string {
	query := u.Query()

	names := make([]string, 0, len(query))
	for name := range query {
		if !excludedFromSignature[name] {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, query[name]...)
	}

	return values
}

// LegacyValues returns the values named by _keys, in the order _keys lists
// them. This is the mod_dims rule. It covers only the parameters a caller
// opts in, so every other parameter stays open to tampering.
func LegacyValues(u *url.URL) []string {
	keys := u.Query().Get("_keys")
	if keys == "" {
		return nil
	}

	var values []string
	for _, key := range strings.Split(keys, ",") {
		if value := u.Query().Get(key); value != "" {
			values = append(values, value)
		}
	}

	return values
}
