// Copyright 2024 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"crop":             CropOperation,
	"resize":           ResizeOperation,
	"strip":            StripMetadataOperation,
	"format":           FormatOperation,
	"quality":          QualityOperation,
	"sharpen":          SharpenOperation,
	"brightness":       BrightnessOperation,
	"flipflop":         FlipFlopOperation,
	"sepia":            SepiaOperation,
	"grayscale":        GrayScaleOperation,
	"autolevel":        AutolevelOperation,
	"invert":           InvertOperation,
	"rotate":           RotateOperation,
	"thumbnail":        ThumbnailOperation,
	"legacy_thumbnail": LegacyThumbnailOperation,
	"gravity":          GravityOperation,
}

func ResizeOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("ResizeOperation", "args", args)

	// Parse Geometry
	var rect imagick.RectangleInfo

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(args, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	slog.Debug("ResizeOperation", "width", rect.Width, "height", rect.Height)

	return mw.ScaleImage(rect.Width, rect.Height)
}

func ThumbnailOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("ThumbnailOperation", "args", args)

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect imagick.RectangleInfo
	var exception imagick.ExceptionInfo

	slog.Info("ThumbnailOperation", "resizedArgs", resizedArgs)

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(resizedArgs, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	slog.Debug("ThumbnailOperation[resize]", "rect", rect)

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	mw.ThumbnailImage(rect.Width, rect.Height)

	if (flags & imagick.PERCENTVALUE) != 0 {
		flags = imagick.ParseGravityGeometry(mw.Image(), args, &rect, &exception)
		if (flags & imagick.ALLVALUES) == 0 {
			return errors.New("parsing thumbnail (crop) geometry failed")
		}

		slog.Debug("ThumbnailOperation[crop]", "rect", rect)
		mw.CropImage(rect.Width, rect.Height, rect.X, rect.Y)
		return mw.SetImagePage(rect.Width, rect.Height, rect.X, rect.Y)
	}

	return nil
}

func CropOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("CropOperation", "args", args)

	/* Replace spaces with '+'. This happens when some user agents inadvertently
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
	flags := imagick.ParseGravityGeometry(mw.Image(), sanitizedArgs, &rect, &exception)
	if (flags & imagick.ALLVALUES) == 0 {
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

	return mw.SharpenImage(geometry.Rho, geometry.Sigma)
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

func GravityOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("GravityOperation", "args", args)

	gravityMap := map[string]imagick.GravityType{
		"northwest": imagick.GRAVITY_NORTH_WEST,
		"north":     imagick.GRAVITY_NORTH,
		"northeast": imagick.GRAVITY_NORTH_EAST,
		"west":      imagick.GRAVITY_WEST,
		"center":    imagick.GRAVITY_CENTER,
		"east":      imagick.GRAVITY_EAST,
		"southwest": imagick.GRAVITY_SOUTH_WEST,
		"south":     imagick.GRAVITY_SOUTH,
		"southeast": imagick.GRAVITY_SOUTH_EAST,
	}

	gravity, ok := gravityMap[strings.ToLower(args)]
	if !ok {
		return errors.New("unknown gravity")
	}

	return mw.SetImageGravity(gravity)
}

func LegacyThumbnailOperation(mw *imagick.MagickWand, args string) error {
	slog.Debug("LegacyThumbnailOperation", "args", args)

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect imagick.RectangleInfo

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(resizedArgs, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	if rect.Width < 200 && rect.Height < 200 {
		mw.ThumbnailImage(rect.Width, rect.Height)
	} else {
		mw.ScaleImage(rect.Width, rect.Height)
	}

	flags = imagick.ParseAbsoluteGeometry(args, &rect)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (crop) geometry failed")
	}

	width := mw.GetImageWidth()
	height := mw.GetImageHeight()
	x := (width / 2) - (rect.Width / 2)
	y := (height / 2) - (rect.Height / 2)

	mw.CropImage(rect.Width, rect.Height, int(x), int(y))
	mw.SetImagePage(rect.Width, rect.Height, int(x), int(y))

	return nil
}
