package dims

import (
	"errors"
	"strconv"
	"strings"

	"github.com/sagikazarmark/slog-shim"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type Operation func(mw *imagick.MagickWand, args string) error

var Operations = map[string]Operation{
	"crop":       CropOperation,
	"resize":     ResizeOperation,
	"strip":      StripMetadataOperation,
	"format":     FormatOperation,
	"quality":    QualityOperation,
	"sharpen":    SharpenOperation,
	"brightness": BrightnessOperation,
	"flipflop":   FlipFlopOperation,
	"sepia":      SepiaOperation,
	"grayscale":  GrayScaleOperation,
	"autolevel":  AutolevelOperation,
	"invert":     InvertOperation,
	"rotate":     RotateOperation,
}

func ResizeOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("ResizeOperation", "args", args)

	// Parse Geometry
	var x int
	var y int
	var width uint
	var height uint

	flags := imagick.ParseMetaGeometry(args, &x, &y, &width, &height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("invalid geometry")
	}

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	slog.Debug("ResizeOperation", "width", width, "height", height)

	return mw.ScaleImage(width, height)
}

func CropOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("CropOperation", "args", args)

	/* Replace spaces with '+'. This happens when some user agents inadvertantly
	 * escape the '+' as %20 which gets converted to a space.
	 *
	 * Example:
	 *
	 * 900x900%20350%200 is '900x900 350 0' which is an invalid, the following code
	 * coverts this to '900x900+350+0'.
	 *
	 */
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")

	// Parse Geometry
	var rect imagick.RectangleInfo
	var exception imagick.ExceptionInfo
	flags := mw.ParseGravityGeometry(sanitizedArgs, &rect, &exception)
	if flags&imagick.ALLVALUES == 0 {
		return errors.New("invalid geometry")
	}

	slog.Debug("CropOperation", "rect", rect)

	if err := mw.CropImage(rect.Width, rect.Height, rect.X, rect.Y); err != nil {
		return err
	}

	return mw.SetImagePage(rect.Width, rect.Height, rect.X, rect.Y)
}

func StripMetadataOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("StripMetadataOperation")

	return mw.StripImage()
}

func FormatOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("FormatOperation", "args", args)

	return mw.SetImageFormat(args)
}

func QualityOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("QualityOperation", "args", args)

	quality, err := strconv.Atoi(args)
	if err != nil {
		return err
	}

	return mw.SetImageCompressionQuality(uint(quality))
}

func SharpenOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("SharpenOperation", "args", args)

	var geometry imagick.GeometryInfo
	flags := imagick.ParseGeometry(args, &geometry)
	if (flags & imagick.SIGMAVALUE) == 0 {
		geometry.Sigma = 1.0
	}

	return mw.SharpenImage(0, 1)
}

func BrightnessOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("BrightnessOperation", "args", args)

	var geometry imagick.GeometryInfo
	imagick.ParseGeometry(args, &geometry)

	return mw.BrightnessContrastImage(geometry.Rho, geometry.Sigma)
}

func FlipFlopOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("FlipFlopOperation", "args", args)

	if args == "horizontal" {
		return mw.FlopImage()
	} else if args == "vertical" {
		return mw.FlipImage()
	}

	return nil
}

func SepiaOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("SepiaOperation", "args", args)

	threshold, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	return mw.SepiaToneImage(threshold * imagick.QUANTUM_RANGE)
}

func GrayScaleOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("GrayScaleOperation", "args", args)

	if args == "true" {
		return mw.SetImageColorspace(imagick.COLORSPACE_GRAY)
	}

	return nil
}

func AutolevelOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("AutolevelOperation", "args", args)

	if args == "true" {
		return mw.AutoLevelImage()
	}

	return nil
}

func InvertOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("InvertOperation", "args", args)

	if args == "true" {
		return mw.NegateImage(false)
	}

	return nil
}

func RotateOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("RotateOperation", "args", args)

	degrees, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	return mw.RotateImage(imagick.NewPixelWand(), degrees)
}
