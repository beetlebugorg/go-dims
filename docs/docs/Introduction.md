---
slug: /
sidebar_position: 1
---

**go-dims** is a fast, lightweight HTTP microservice for real-time image processing, written in Go
and powered by [libvips](https://libvips.github.io/libvips/). It’s a modern, drop-in replacement for
[mod_dims](https://github.com/beetlebugorg/mod_dims), fully compatible with the DIMS API.

Designed for use in websites, publishing platforms, and CDN-backed applications, `go-dims` helps you
generate image variants on-the-fly — without the need for precomputing or storing multiple
renditions.

## ✨ Features
- ✅ Resize, crop, rotate, flip, grayscale, and more
- ✅ Add watermarks
- ✅ Strip metadata, control quality, convert formats
- ✅ Export in JPEG, PNG, and WebP
- ✅ Sign and validate requests for secure public access
- ✅ Load images from file, a URL, or S3
- ✅ Deploy as a Docker image, or AWS Lambda function
- ✅ Drop-in replacement for [mod_dims](https://github.com/beetlebugorg/mod_dims)

## Why use go-dims?

**💡 On-demand image transformation**

- Resize, crop, convert formats, and more — all defined via URL.
- Built-in support for signatures to prevent misuse and ensure cacheability.
- Avoid bloated image storage by rendering variants only when requested.

**⚡ Fast, minimal, portable**

- Built on [libvips](https://libvips.github.io/libvips/), the fastest image processing library.
- Single static binary with zero runtime dependencies.
- Docker image is just **~11MB** compressed.
- Built for **linux/amd64** and **linux/arm64**.

**🔒 Secure**
- Clean, HMAC-SHA256 signed URLs to ensure safe, tamperproof transformations.
- The go-dims Docker image is built from scratch with a statically compiled binary, no shell, no package manager, and no extraneous libraries — minimizing the attack surface.

**🛠 Developer-friendly**
- Easily define image variants directly in frontend code or templates.
- Uses minimal memory — perfect for local development.
- Runs equally well locally, in containers, or as an AWS Lambda function.

## 📦 Deployment Options
- Static binary: Just download and run.
- Standalone Docker: Launch anywhere in seconds.
- AWS Lambda: Compile and deploy as a fast, small Lambda function.

## License

go-dims is licensed under the MIT license. See the [LICENSE](LICENSE) file for details.

For software used by this project, and their licenses, see the [NOTICE](NOTICE) file.