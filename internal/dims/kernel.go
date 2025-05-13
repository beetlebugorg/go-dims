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
	"github.com/beetlebugorg/go-dims/internal/gox/imagex/colorx"
	"github.com/davidbyttow/govips/v2/vips"
)

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
	HttpRequest            http.Request
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

	shrinkFactor int
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

func (r *Request) LoadImage(sourceImage *core.Image) (*vips.ImageRef, error) {
	r.SourceImage = *sourceImage

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

	opts.WebpExportParams.StripMetadata = stripMetadata
	opts.WebpExportParams.ReductionEffort = 6

	opts.PngExportParams.StripMetadata = stripMetadata
	opts.GifExportParams.StripMetadata = stripMetadata
	opts.TiffExportParams.StripMetadata = stripMetadata

	for _, command := range r.Commands() {
		region := trace.StartRegion(ctx, command.Name)

		if operation, ok := VipsTransformCommands[command.Name]; ok {
			if command.Name == "crop" {
				adjustedArgs, err := adjustCropAfterShrink(command.Args, r.shrinkFactor)
				if err != nil {
					return "", nil, err
				}

				command.Args = adjustedArgs
			}

			if command.Name == "strip" && command.Args == "false" {
				stripMetadata = false
			}

			if err := operation(image, command.Args); err != nil && !errorImage {
				return "", nil, err
			}
		} else if operation, ok := VipsExportCommands[command.Name]; ok && !errorImage {
			if err := operation(image, command.Args, &opts); err != nil {
				return "", nil, err
			}
		} else if operation, ok := VipsRequestCommands[command.Name]; ok && !errorImage {
			if err := operation(image, command.Args, operations.RequestOperation{
				Request: r.HttpRequest,
				Config:  r.Config,
			}); err != nil {
				return "", nil, err
			}
		}

		region.End()
	}

	if stripMetadata {
		image.RemoveMetadata()
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

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Set etag header.
	if r.SourceImage.Etag != "" {
		var h hash.Hash
		if r.Config.EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.Id))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		}

		etag := fmt.Sprintf("%x", h.Sum(nil))

		w.Header().Set("ETag", etag)
	}

	if r.SourceImage.LastModified != "" {
		w.Header().Set("Last-Modified", r.SourceImage.LastModified)
	}
}

func (r *Request) SendImage(w http.ResponseWriter, status int, imageFormat string, imageBlob []byte) error {
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

func (r *Request) SendError(w http.ResponseWriter, err error) error {
	message := err.Error()

	// Strip stack from vips errors.
	if strings.HasPrefix(message, "VipsOperation:") {
		message = message[0:strings.Index(message, "\n")]
	}

	slog.Error("SendError", "message", message)

	// Set status code.
	status := http.StatusInternalServerError
	var fetchError *core.FetchError
	if errors.As(err, &fetchError) {
		status = fetchError.Status
	}

	errorImage, err := vips.Black(512, 512)
	if err != nil {
		return err
	}

	if err := errorImage.BandJoinConst([]float64{0, 0}); err != nil {
		return err
	}

	backgroundColor, err := colorx.ParseHexColor(r.Config.Error.Background)
	if err != nil {
		return err
	}

	red, green, blue, _ := backgroundColor.RGBA()
	redI := float64(red) / 65535 * 255
	greenI := float64(green) / 65535 * 255
	blueI := float64(blue) / 65535 * 255

	if err := errorImage.Linear([]float64{0, 0, 0}, []float64{redI, greenI, blueI}); err != nil {
		return err
	}

	imageType, imageBlob, err := r.ProcessImage(errorImage, true)
	if err != nil {
		// If processing failed because of a bad command then return the image as-is.
		exportOptions := vips.NewJpegExportParams()
		exportOptions.Quality = 1
		imageBytes, _, _ := errorImage.ExportJpeg(exportOptions)

		return r.SendImage(w, status, "jpg", imageBytes)
	}

	return r.SendImage(w, status, imageType, imageBlob)
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

func adjustCropAfterShrink(args string, factor int) (string, error) {
	rect, err := geometry.ParseGeometry(args)
	if err != nil {
		return "", err
	}

	rect.X = int(float64(rect.X) / float64(factor))
	rect.Y = int(float64(rect.Y) / float64(factor))

	if rect.Width > 0 {
		rect.Width = float64(rect.Width) / float64(factor)
	}

	if rect.Height > 0 {
		rect.Height = float64(rect.Height) / float64(factor)
	}

	// Output full geometry
	if rect.Y > 0 {
		return fmt.Sprintf("%dx%d+%d+%d", int(rect.Width), int(rect.Height), int(rect.X), int(rect.Y)), nil
	}

	// Output geometry without Y
	if rect.X > 0 {
		return fmt.Sprintf("%dx%d+%d", int(rect.Width), int(rect.Height), int(rect.X)), nil
	}

	// Output geometry without offsets
	if rect.Height > 0 {
		return fmt.Sprintf("%dx%d", int(rect.Width), int(rect.Height)), nil
	}

	return fmt.Sprintf("%dx", int(rect.Width)), nil
}
