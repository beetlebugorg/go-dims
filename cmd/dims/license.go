// Copyright 2025 Jeremy Collins. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	root "github.com/beetlebugorg/go-dims"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type LicenseInfo struct {
	Project string `json:"project"`
	SPDX    string `json:"spdx"`
	File    string `json:"file"`
}

// LicenseCmd is a command to show a list of licenses used by go-dims.
type LicenseCmd struct {
	License string `arg:"" optional:"" help:"Name of the project to show the license for."`
	Pager   bool   `default:"true" help:"Show license details."`
}

func (cmd *LicenseCmd) Run() error {
	if cmd.License != "" {
		return cmd.Show()
	}

	err := listLicenses()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *LicenseCmd) Show() error {
	licenses, err := loadLicenseMetadata()
	if err != nil {
		return err
	}

	var builder strings.Builder

	for _, l := range licenses {
		if strings.EqualFold(l.Project, cmd.License) {
			content, err := root.LicenseFS.ReadFile(l.File)
			if err != nil {
				return err
			}

			builder.WriteString(fmt.Sprintf("License for %s (%s):\n\n", l.Project, l.SPDX))
			builder.WriteString(fmt.Sprintln(string(content)))

			pageOutput([]byte(builder.String()))

			return nil
		}
	}
	return errors.New("license not found")
}

func listLicenses() error {
	licenses, err := loadLicenseMetadata()
	if err != nil {
		return err
	}
	sort.Slice(licenses, func(i, j int) bool {
		return licenses[i].Project < licenses[j].Project
	})

	fmt.Printf("%-40s  %s\n", "Project", "License")
	fmt.Printf("%-40s  %s\n", "------", "------")
	for _, l := range licenses {
		fmt.Printf("%-40s  %s\n", l.Project, l.SPDX)
	}
	return nil
}

func loadLicenseMetadata() ([]LicenseInfo, error) {
	data, err := root.LicenseFS.ReadFile("LICENSES/index.json")
	if err != nil {
		return nil, err
	}
	var licenses []LicenseInfo
	if err := json.Unmarshal(data, &licenses); err != nil {
		return nil, err
	}
	return licenses, nil
}

func pageOutput(data []byte) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		_, _ = os.Stdout.Write(data)
		return nil
	}

	// Check if pager is available
	// Verify the pager is available on PATH
	path, err := exec.LookPath(pager)
	if err != nil {
		// Fallback: just print to stdout
		_, err := os.Stdout.Write(data)
		return err
	}

	// Try to start pager
	cmd := exec.Command(path)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// If pager fails, fallback to stdout
	if err := cmd.Run(); err != nil {
		_, _ = os.Stdout.Write(data)
		return nil
	}
	return nil
}
