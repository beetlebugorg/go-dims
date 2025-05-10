package dims

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/beetlebugorg/go-dims/internal/dims/core"
	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/beetlebugorg/go-dims/internal/dims/operations"
	"github.com/davidbyttow/govips/v2/vips"
)

type Kernel interface {
	ValidateSignature() bool
	FetchImage() error
	ProcessImage() (string, []byte, error)
	ProcessCommand(ctx context.Context, command operations.Command) error
	SendHeaders(w http.ResponseWriter)
	SendImage(w http.ResponseWriter, status int, imageType string, imageBlob []byte) error
	SendError(w http.ResponseWriter, status int, message string)
}

type ErrorImageGenerator interface {
	GenerateErrorImage(w http.ResponseWriter, status int, message string)
}

type ImageProcessor interface {
	ProcessImage() (string, []byte, error)
}

type Signer interface {
	ValidateSignature(request Request) bool
	Sign(request Request) string
}

type Request struct {
	HttpRequest            *http.Request
	Id                     string      // The hash of the request -> hash(clientId + commands + imageUrl).
	Signature              string      // The signature of the request.
	Config                 core.Config // The global configuration.
	ClientId               string      // The client ID of this request.
	ImageUrl               string      // The image URL that is being manipulated.
	SendContentDisposition bool        // The content disposition of the request.
	RawCommands            string      // The commands ('resize/100x100', 'strip/true/format/png', etc).
	Error                  bool        // Whether the error image is being served.
	Timestamp              int64       // The timestamp of the request
	SourceImage            core.Image  // The source image.
}

var VipsTransformCommands = map[string]operations.VipsTransformOperation{
	"crop":             operations.CropCommand,
	"resize":           operations.ResizeCommand,
	"sharpen":          operations.SharpenCommand,
	"brightness":       operations.BrightnessCommand,
	"flipflop":         operations.FlipFlopCommand,
	"sepia":            operations.SepiaCommand,
	"grayscale":        operations.GrayscaleCommand,
	"autolevel":        operations.AutolevelCommand,
	"invert":           operations.InvertCommand,
	"rotate":           operations.RotateCommand,
	"thumbnail":        operations.ThumbnailCommand,
	"legacy_thumbnail": operations.LegacyThumbnailCommand,
}

var VipsExportCommands = map[string]operations.VipsExportOperation{
	"strip":   operations.StripMetadataCommand,
	"format":  operations.FormatCommand,
	"quality": operations.QualityCommand,
}

var VipsRequestCommands = map[string]operations.VipsRequestOperation{
	"watermark": operations.Watermark,
}

// FetchImage downloads the image from the given URL.
func (r *Request) FetchImage() error {
	slog.Info("downloadImage", "url", r.ImageUrl)

	timeout := time.Duration(r.Config.Timeout.Download) * time.Millisecond
	sourceImage, err := core.FetchImage(r.ImageUrl, timeout)
	if err != nil {
		return err
	}

	if sourceImage.Status != 200 {
		return fmt.Errorf("failed to download image: %d", sourceImage.Status)
	}

	r.SourceImage = *sourceImage

	return nil
}

// ProcessImage will execute the commands on the image.
func (r *Request) ProcessImage() (string, []byte, error) {
	slog.Debug("executeVips")

	image, err := vips.NewImageFromBuffer(r.SourceImage.Bytes)
	if err != nil {
		return "", nil, err
	}
	importParams := vips.NewImportParams()
	importParams.AutoRotate.Set(true)

	requestedSize, err := r.requestedImageSize()
	if err == nil && vips.DetermineImageType(r.SourceImage.Bytes) == vips.ImageTypeJPEG {
		xs := image.Width() / int(requestedSize.Width)
		ys := image.Height() / int(requestedSize.Height)

		if (xs > 2) || (ys > 2) {
			importParams.JpegShrinkFactor.Set(4)
		}
	}

	image, err = vips.LoadImageFromBuffer(r.SourceImage.Bytes, importParams)
	if err != nil {
		return "", nil, err
	}

	ctx := context.Background()

	slog.Info("executeVips", "image", image, "buffer-size", len(r.SourceImage.Bytes))

	// Execute the commands.
	ctx, task := trace.NewTask(ctx, "v5.ProcessImage")
	defer task.End()

	opts := operations.ExportOptions{
		ImageType:        image.Format(),
		JpegExportParams: vips.NewJpegExportParams(),
		PngExportParams:  vips.NewPngExportParams(),
		WebpExportParams: vips.NewWebpExportParams(),
		GifExportParams:  vips.NewGifExportParams(),
		TiffExportParams: vips.NewTiffExportParams(),
	}

	stripMetadata := r.Config.StripMetadata
	opts.JpegExportParams.StripMetadata = stripMetadata
	opts.JpegExportParams.SubsampleMode = vips.VipsForeignSubsampleAuto
	opts.JpegExportParams.OptimizeCoding = true
	opts.JpegExportParams.OptimizeScans = true
	opts.JpegExportParams.TrellisQuant = true

	opts.PngExportParams.StripMetadata = stripMetadata
	opts.WebpExportParams.StripMetadata = stripMetadata
	opts.GifExportParams.StripMetadata = stripMetadata
	opts.TiffExportParams.StripMetadata = stripMetadata

	for _, command := range r.Commands() {
		region := trace.StartRegion(ctx, command.Name)

		if operation, ok := VipsTransformCommands[command.Name]; ok {
			slog.Debug("executeTransformCommand", "command", command.Name, "args", command.Args)
			if err := operation(image, command.Args); err != nil {
				return "", nil, err
			}
		} else if operation, ok := VipsExportCommands[command.Name]; ok {
			slog.Debug("executeExportCommand", "command", command.Name, "args", command.Args)
			if err := operation(image, command.Args, &opts); err != nil {
				return "", nil, err
			}
		} else if operation, ok := VipsRequestCommands[command.Name]; ok {
			slog.Debug("executeRequestCommand", "command", command.Name, "args", command.Args)
			if err := operation(image, command.Args, operations.RequestOperation{
				Request: *r.HttpRequest,
				Config:  r.Config,
			}); err != nil {
				return "", nil, err
			}
		}

		region.End()
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

	return vips.ImageTypes[image.Format()], imageBytes, nil
}

func (r *Request) SendHeaders(w http.ResponseWriter) {
	maxAge := r.Config.OriginCacheControl.Default
	edgeControlTtl := r.Config.EdgeControl.DownstreamTtl

	if r.Config.OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			minCacheAge := r.Config.OriginCacheControl.Min
			if minCacheAge != 0 && maxAge <= minCacheAge {
				maxAge = minCacheAge
			}

			// If above maximum, set to maximum.
			maxCacheAge := r.Config.OriginCacheControl.Max
			if maxCacheAge != 0 && maxAge >= maxCacheAge {
				maxAge = maxCacheAge
			}
		}
	}

	if r.Error {
		maxAge = r.Config.OriginCacheControl.Error
	}

	// Set cache headers.
	if maxAge > 0 {
		slog.Debug("sendImage", "maxAge", maxAge)

		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		w.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
	}

	if edgeControlTtl > 0 {
		w.Header().Set("Edge-Control", fmt.Sprintf("downstream-ttl=%d", edgeControlTtl))
	}

	// Set content disposition.
	if r.SendContentDisposition {
		// Grab filename from imageUrl
		u, err := url.Parse(r.ImageUrl)
		if err != nil {
			return
		}

		filename := filepath.Base(u.Path)

		slog.Debug("sendImage", "sendContentDisposition", r.SendContentDisposition, "filename", filename)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Set etag header.
	if r.SourceImage.Etag != "" || r.SourceImage.LastModified != "" {
		var h hash.Hash
		if r.Config.Signing.EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.Id))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		} else if r.SourceImage.LastModified != "" {
			h.Write([]byte(r.SourceImage.LastModified))
		}

		etag := fmt.Sprintf("%x", h.Sum(nil))

		w.Header().Set("Etag", etag)
	}
}

func (r *Request) SendImage(w http.ResponseWriter, status int, imageFormat string, imageBlob []byte) error {
	slog.Info("SendImage", "status", status, "format", imageFormat, "size", len(imageBlob))

	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	// Set content type.
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageFormat)))

	// Set content length
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	w.WriteHeader(status)

	// Write the image.
	_, err := w.Write(imageBlob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) SendError(w http.ResponseWriter, status int, message string) {
	slog.Info("sendError", "status", status, "message", message)
}

func (r *Request) Commands() []operations.Command {
	commands := make([]operations.Command, 0)
	parsedCommands := strings.Split(strings.Trim(r.RawCommands, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, operations.Command{
			Name: command,
			Args: args,
		})

		slog.Debug("parsedCommand", "command", command, "args", args)
	}

	return commands
}

func sourceMaxAge(header string) (int, error) {
	if header == "" {
		return 0, errors.New("empty header")
	}

	pattern, _ := regexp.Compile(`max-age=(\d+)`)
	match := pattern.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0, errors.New("unable to convert to int")
		}

		return sourceMaxAge, nil
	}

	return 0, errors.New("max-age not found in header")
}

// Parse through the requested commands and return requested image size for thumbnail and resize
// operations.
//
// This is used while reading an image to improve performance when generating thumbnails from very
// large images.
func (r *Request) requestedImageSize() (geometry.Geometry, error) {
	for _, command := range r.Commands() {
		if command.Name == "thumbnail" || command.Name == "resize" {
			var rect = geometry.ParseGeometry(command.Args)

			if rect.Width > 0 && rect.Height > 0 {
				return rect, nil
			}

		}
	}

	return geometry.Geometry{}, errors.New("no resize or thumbnail command found")
}
