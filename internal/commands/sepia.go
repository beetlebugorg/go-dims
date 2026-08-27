package commands

import "github.com/davidbyttow/govips/v2/vips"

// SepiaCommand is not implemented yet. mod_dims applies MagickSepiaToneImage
// with the argument as a threshold, and this returns the image untouched.
// See the tracking issue before relying on it.
func SepiaCommand(image *vips.ImageRef, args string) error {
	return nil
}
