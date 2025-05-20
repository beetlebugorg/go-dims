---
sidebar_position: 3
---

# Image Compression

These settings control how `go-dims` compresses output images across supported formats: JPEG, PNG, and WebP.

Defaults are tuned for a balance of quality, size, and performance. If you have stricter performance, bandwidth, or quality goals, you can override them here.

---

## JPEG Compression

### `DIMS_JPEG_QUALITY`

Controls overall JPEG image quality (1–100).

- **Default:** `80`

A higher value improves quality at the cost of file size.

---

### `DIMS_JPEG_INTERLACE`

Enables progressive (interlaced) JPEGs, which load in multiple passes.

- **Default:** `false`

---

### `DIMS_JPEG_OPTIMIZE_CODING`

Enables Huffman table optimization for slightly smaller files.

- **Default:** `true`

---

### `DIMS_JPEG_SUBSAMPLE_MODE`

Controls chroma subsampling (reducing color detail to save space).

- **Default:** `true` (subsampling enabled)

---

### `DIMS_JPEG_TRELLIS_QUANT`

Enables trellis quantization for better compression at same quality.

- **Default:** `false`

More CPU-intensive, but can reduce file size slightly.

---

### `DIMS_JPEG_OVERSHOOT_DERINGING`

Improves visual quality in sharp transitions (e.g., text or edges).

- **Default:** `false`

May improve detail slightly in compressed images.

---

### `DIMS_JPEG_OPTIMIZE_SCANS`

Enables scan optimization (more efficient progressive JPEG encoding).

- **Default:** `false`

Reduced file size at the cost of increased compression time.

---

### `DIMS_JPEG_QUANT_TABLE`

Selects a predefined quantization table.

- **Default:** `3`

See [libvips](https://www.libvips.org/API/current/VipsForeignSave.html#vips-jpegsave) documentation for available options.

---

## PNG Compression

### `DIMS_PNG_INTERLACE`

Enables Adam7 interlacing (progressive PNG).

- **Default:** `false`

---

### `DIMS_PNG_COMPRESSION`

Compression level from `0` (fastest) to `9` (best compression).

- **Default:** `4`

Higher values compress more but take longer to encode.

---

## WebP Compression

### `DIMS_WEBP_QUALITY`

Controls visual quality for WebP output (1–100).

- **Default:** `80`

---

### `DIMS_WEBP_COMPRESSION`

Selects WebP compression mode.

- **Default:** `lossy`
- Options: `lossy`, `lossless`

---

### `DIMS_WEBP_REDUCTION_EFFORT`

Effort level from `0` (fastest) to `6` (slowest, best compression).

- **Default:** `4`

Use higher values to reduce size at the cost of CPU usage.