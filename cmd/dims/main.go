//go:build !lambda

package main

import (
	"os"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Serve   ServeCmd      `cmd:"" help:"Runs the DIMS service."`
	Encrypt EncryptionCmd `cmd:"" help:"Encrypt an eurl."`
	Decrypt DecryptionCmd `cmd:"" help:"Decrypt an eurl."`
	Health  HealthCmd     `cmd:"" help:"Check the health of the DIMS service."`
	Sign    SignCmd       `cmd:"" help:"Sign an image URL."`
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run()
	if err != nil {
		os.Exit(1)
	}
}
