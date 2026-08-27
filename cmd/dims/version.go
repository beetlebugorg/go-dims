package main

import (
	"fmt"

	"github.com/beetlebugorg/go-dims/internal/core"
)

type VersionCmd struct {
}

func (v *VersionCmd) Run() error {
	fmt.Println(core.Version)

	return nil
}
