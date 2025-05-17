# go-dims

**go-dims** is a fast, lightweight HTTP microservice for real-time image processing, written in Go
and powered by [libvips](https://libvips.github.io/libvips/). It’s a modern, drop-in replacement for
[mod_dims](https://github.com/beetlebugorg/mod_dims), fully compatible with the DIMS API.

Designed for use in websites, publishing platforms, and CDN-backed applications, `go-dims` helps you
generate image variants on-the-fly — without the need for precomputing or storing multiple
renditions.

## Features
- ✅ Resize, crop, rotate, flip, grayscale, and more
- ✅ Add watermarks
- ✅ Strip metadata, control quality, convert formats
- ✅ Export in JPEG, PNG, and WebP
- ✅ Support for both dims4 and legacy URL structures
- ✅ Sign and validate requests for secure public access
- ✅ Load images from file, a URL, or S3
- ✅ Deploy as a Docker image, or AWS Lambda function

## Why use go-dims?

**💡 On-demand image transformation**

- Resize, crop, convert formats, and more — all defined via URL.
- Built-in support for signatures to prevent misuse and ensure cacheability.
- Avoid bloated image storage by rendering variants only when requested.

**⚡ Fast, minimal, portable**

- Built on [libvips](https://libvips.github.io/libvips/), the fastest image processing library.
- Single static binary with zero runtime dependencies.
- Docker image is just **~9MB** compressed.
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
 
---

## Getting Started

Run locally in development mode (no signature required):

```bash
docker run \
  -e DIMS_DEVELOPMENT_MODE=true \
  -e DIMS_DEBUG_MODE=true \
  -e DIMS_SIGNING_KEY=devmode \
  -p 8080:8080 \
  ghcr.io/beetlebugorg/go-dims serve
```

Then, open your browser and navigate to:

```
http://localhost:8080/v5/resize/200x200/?url=pexels-photo-1539116.jpeg
```

## 🧩 Supported Transformations

| Type          | Command                              | Example              |
|---------------|--------------------------------------|----------------------|
| Resize        | `resize/WxH`                         | `resize/300x300`     |
| Crop          | `crop/WxH+X+Y`                       | `crop/200x100+0+25%25` |
| Thumbnail     | `thumbnail/WxH`                      | `thumbnail/200x200`  |
| Watermark     | `watermark/<args>`                   | `watermark/1,.35,se` |
| Format        | `format/<string>`                    | `format/webp`        |
| Strip         | `strip/<bool>`                       | `strip/true`         |
| Quality       | `quality/<0-100>`                    | `quality/85`         |
| Rotate        | `rotate/<0-360>`                     | `rotate/90`          |
| Sharpen       | `sharpen/TxS`                        | `sharpen/1x2`        |
| Autolevel     | `autolevel/<bool>`                   | `autolevel/true`     |
| Brightness    | `brightness/<brightness>x<contrast>` | `brightness/5x25`    |
| Flip/Flop     | `flipflop/<horizontal\|vertical>`    | `flip/horizontal`    |
| Invert        | `invert/<bool>`                      | `invert/true`        |

## Examples

### **Original**

<img src="./resources/pexels-photo-1539116.jpeg" alt="Example" width="300">

### Thumbnail, then sharpen

Thumbnail resize to 200x200 pixels, keeping the aspect ratio and cropping the image if necessary.

```
http://localhost:8080/v5/thumbnail/200x200/sharpen/1x2/?url=pexels-photo-1539116.jpeg
```

![Thumbnail](./resources/readme/thumbnail.jpg)

### Resize with center crop

```
http://localhost:8080/v5/resize/200x200/crop/200x100+0+25%25/sharpen/1x2?url=pexels-photo-1539116.jpeg
```

![Resize with center crop](./resources/readme/resize-center-crop.jpg)

### **Watermark**

Add a watermark to the bottom right corner of the image.

```
http://localhost:8080/v5/thumbnail/200x200/sharpen/1x2/watermark/1,.35,se?url=pexels-photo-1539116.jpeg&overlay=hex-lab.png
```

![Watermark](./resources/readme/watermark.jpg)

## 🔑 Generating a Signed URL

Use the built-in CLI to generate secure, signed URLs:

```bash
❯ ./build/dims sign --key-file=signing-key.txt "http://localhost:8080/v5/resize/200x200/?url=pexels-photo-1539116.jpeg"

http://localhost:8080/v5/resize/200x200/?sig=33032505c0f4b3d43674b49575d9e379470ac6d7e7fa3e055b248802ee6867&url=pexels-photo-1539116.jpeg
