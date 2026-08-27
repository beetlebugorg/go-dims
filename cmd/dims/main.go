//go:build !lambda

package main

import (
	"github.com/alecthomas/kong"
)

var CLI struct {
	Serve   ServeCmd      `cmd:"" help:"Runs the DIMS service."`
	Encrypt EncryptionCmd `cmd:"" help:"Encrypt an eurl."`
	Decrypt DecryptionCmd `cmd:"" help:"Decrypt an eurl."`
	Health  HealthCmd     `cmd:"" help:"Check the health of the DIMS service."`
	Sign    SignCmd       `cmd:"" help:"Sign an image URL."`
	Version VersionCmd    `cmd:"" help:"Print the version."`
}

func main() {
	ctx := kong.Parse(&CLI)
	ctx.FatalIfErrorf(ctx.Run())
}
