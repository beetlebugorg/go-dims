#!/usr/bin/env bash
# Print the Homebrew formula for a released version, reading each archive's
# sha256 out of <dist-dir>, the zips the release workflow just built. The tap
# job pipes this into Formula/go-dims.rb.
#
# Usage: brew-formula.sh <version> <dist-dir>
set -euo pipefail

version="$1"
dist="$2"

# The repository the release lives in, so a fork's test release yields a
# formula pointing at the fork's own downloads.
repo="${GITHUB_REPOSITORY:-beetlebugorg/go-dims}"
base="https://github.com/$repo/releases/download/v$version"

sha() { shasum -a 256 "$dist/dims-$1.zip" | cut -d' ' -f1; }

# Resolved before the heredoc. A command substitution that fails inside one is
# not caught by set -e, and a missing archive would emit an empty sha256.
macos_arm="$(sha macos-arm64)"
macos_intel="$(sha macos-amd64)"
linux_arm="$(sha linux-arm64)"
linux_intel="$(sha linux-amd64)"

cat <<EOF
class GoDims < Formula
  desc "On-the-fly image resizing server, with a command to sign image URLs"
  homepage "https://github.com/beetlebugorg/go-dims"
  version "$version"
  license "MIT"

  on_macos do
    # The macOS build links Homebrew's libvips. The Linux build is static and
    # needs nothing at run time.
    depends_on "vips"

    on_arm do
      url "$base/dims-macos-arm64.zip"
      sha256 "$macos_arm"
    end
    on_intel do
      url "$base/dims-macos-amd64.zip"
      sha256 "$macos_intel"
    end
  end

  on_linux do
    on_arm do
      url "$base/dims-linux-arm64.zip"
      sha256 "$linux_arm"
    end
    on_intel do
      url "$base/dims-linux-amd64.zip"
      sha256 "$linux_intel"
    end
  end

  def install
    bin.install "dims"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dims version")
  end
end
EOF
