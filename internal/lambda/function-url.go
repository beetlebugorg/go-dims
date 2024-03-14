package lambda

import (
	"crypto/sha256"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"log/slog"
	"net/url"
	"strings"
)

var CommandsLambda = map[string]dims.Operation{
	"crop":       v4.CropCommand,
	"resize":     v4.ResizeCommand,
	"strip":      v4.StripMetadataCommand,
	"format":     v4.FormatCommand,
	"quality":    v4.QualityCommand,
	"sharpen":    v4.SharpenCommand,
	"brightness": v4.BrightnessCommand,
	"flipflop":   v4.FlipFlopCommand,
	"sepia":      v4.SepiaCommand,
	"grayscale":  v4.GrayscaleCommand,
	"autolevel":  v4.AutolevelCommand,
	"invert":     v4.InvertCommand,
	"rotate":     v4.RotateCommand,
	"thumbnail":  v4.ThumbnailCommand,
	"gravity":    v5.GravityCommand,
}

type RequestLambdaFunctionURL struct {
	v5.RequestV5
}

func NewLambdaFunctionURLRequest(event events.LambdaFunctionURLRequest, config dims.Config) *RequestLambdaFunctionURL {
	u, err := url.Parse(event.RawPath + "?" + event.RawQueryString)
	if err != nil {
		return nil
	}

	slog.Info("NewLambdaFunctionURLRequest", "event", event)

	// /v5/dims/{commands...}
	rawCommands := strings.TrimLeft(event.RawPath, "/v5/dims/")

	h := sha256.New()
	h.Write([]byte(u.Query().Get("clientId")))
	h.Write([]byte(rawCommands))
	h.Write([]byte(u.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestLambdaFunctionURL{
		RequestV5: v5.RequestV5{
			Request: dims.Request{
				Id:        requestHash,
				Config:    config,
				ClientId:  u.Query().Get("clientId"),
				ImageUrl:  u.Query().Get("url"),
				Commands:  dims.ParseCommands(rawCommands, CommandsLambda),
				Signature: u.Query().Get("sig"),
			},
		},
	}
}
