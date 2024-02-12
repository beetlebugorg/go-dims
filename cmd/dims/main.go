package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alecthomas/kong"
	"github.com/beetlebugorg/go-dims/pkg/dims"
)

type ServeCmd struct {
	Bind  string `help:"Bind address to serve on." default:"127.0.0.1:8080"`
	Debug bool   `help:"Enable debug mode."`
}

func (s *ServeCmd) Run() error {
	err := http.ListenAndServe(s.Bind, dims.NewHandler())
	if err != nil {
		slog.Error("Server failed.", "error", err)
		return err
	}
	return nil
}

type SignCmd struct {
	Timestamp string `arg:"" name:"timestamp" help:"Expiration timestamp" type:"timestamp"`
	Secret    string `arg:"" name:"secret" help:"Secret key."`
	Commands  string `arg:"" name:"commands" help:"Commands to sign."`
	ImageURL  string `arg:"" name:"image_url" help:"Image URL to sign."`
}

func (s *SignCmd) Run() error {
	hash := dims.Sign(s.Timestamp, s.Secret, s.Commands, s.ImageURL)
	fmt.Printf("%s\n", hash)
	return nil
}

var CLI struct {
	Serve ServeCmd `cmd:"" help:"Runs the DIMS service."`
	Sign  SignCmd  `cmd:"" help:"Signs the given image URL."`
}

func main() {
	ctx := kong.Parse(&CLI)
	ctx.Run()
}
