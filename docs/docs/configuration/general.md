---
sidebar_position: 1
---

# General Settings

These settings control runtime behavior, error image styling, default output formats, and metadata handling. Most can be left at their defaults, but a few — like the error background color — are worth customizing for your environment.

---

## `DIMS_ERROR_BACKGROUND`

Sets the background color of the error image.

- **Default:** `#5adafd`

This color will appear when an error image is generated, which can happen when:
- The source image could not be downloaded
- The image could not be processed (e.g., invalid format or size)
- The server is under memory or resource pressure

Think of it as the visual "error signal" for `go-dims`. When you see this image, it’s usually a sign to check your logs.

:::note 

mod_dims allowed setting a `NOIMAGE` fallback file — `go-dims` no longer supports that. Error images are now generated dynamically using libvips.

:::

---

## `DIMS_STRIP_METADATA`

Enables or disables automatic removal of image metadata (EXIF, profiles, etc.).

- **Default:** `true`

Stripping metadata is recommended for privacy, security, and cache efficiency. However, in use cases like DAM (Digital Asset Management), you may want to retain this data.

To disable globally:

```
DIMS_STRIP_METADATA=false
```

You can also override this per-request using the [`strip`](../operations/output/strip.md) command.

---

## `DIMS_INCLUDE_DISPOSITION`

Controls whether to include the `Content-Disposition` header in image responses.

- **Default:** `false`

When enabled, the response includes:

```
Content-Disposition: inline; filename=<filename>
```

To trigger downloads instead of inline display, append `download=1` to the request URL. That changes the header to:

```
Content-Disposition: attachment; filename=<filename>
```

`download=1` also sends the header on its own, whether or not this setting is enabled.

The filename comes from the last path segment of the source URL. It is quoted and encoded, and the header is omitted when the name cannot be represented safely.

---

## `DIMS_DOWNLOAD_TIMEOUT`

Sets the maximum time (in milliseconds) an origin image download is allowed before being cancelled.

- **Default:** `3000`

Example:
```
DIMS_DOWNLOAD_TIMEOUT=5000
```

---

## Request Cost Limits

Processing cost is close to linear in the number of pixels produced. Measured on one thread with libvips 8.18:

| Output | Pixels | Time |
|---|---|---|
| 1000x1000 | 1 MP | 35 ms |
| 4000x4000 | 16 MP | 0.5 s |
| 7000x7000 | 49 MP | 1.6 s |
| 10000x10000 | 100 MP | 3.2 s |
| 14000x14000 | 196 MP | 6.5 s |

That is roughly 30 megapixels per second per thread, so a pixel limit bounds how long one request can occupy a worker.

### `DIMS_MAX_OUTPUT_PIXELS`

The largest image the service will produce, in pixels.

- **Default:** `50000000` (50 MP, about 7000x7000, roughly 1.6 seconds)

Checked after every command, since a command can grow the image. A request above the limit returns `400` before any pixels are computed, because width and height are metadata.

```
DIMS_MAX_OUTPUT_PIXELS=16000000
```

### `DIMS_MAX_SOURCE_PIXELS`

The largest source image the service will accept, in pixels.

- **Default:** `100000000` (100 MP, about 10000x10000)

Checked once the source header is read. This mainly guards against a small file that declares enormous dimensions, since a large source reduced to a thumbnail is cheap thanks to shrink on load.

Set either to `0` to disable that check.

### `DIMS_MAX_SOURCE_BYTES`

The largest source image the service will download, in bytes.

- **Default:** `67108864` (64 MB)

A `Content-Length` above the limit is refused before the body is read. A body that passes the limit during the read is refused as well. Both return `400`.

The limit applies to every source backend, and to the `overlay` image the [watermark](../operations/special/watermark.md) command downloads.

Set it to `0` to disable the check.

```
DIMS_MAX_SOURCE_BYTES=16777216
```

### `DIMS_MAX_CONCURRENT`

How many images may be processed at once.

- **Default:** `0`, which derives a value from the CPU count

The pixel limits above bound what one request may ask for. This bounds how many may ask at the same time. Without it, a burst of individually reasonable requests still saturates the machine and every one of them slows down together.

Set a positive number to choose the limit directly, or `-1` to remove it.

```
DIMS_MAX_CONCURRENT=8
```

The slot is taken after the source image has been downloaded, since a request waiting on an origin is not competing for CPU.

### `DIMS_MAX_DOWNLOAD_CONCURRENT`

How many source images may be downloaded at once.

- **Default:** `0`, which derives a value from the CPU count

A download takes a slot from this limit. Processing takes a slot from `DIMS_MAX_CONCURRENT`.

Set a positive number to choose the limit directly, or `-1` to remove it.

```
DIMS_MAX_DOWNLOAD_CONCURRENT=16
```

### `DIMS_MAX_CONCURRENT_WAIT`

How long a request queues for a slot before it is refused, in milliseconds. It applies to both limits above.

- **Default:** `5000`

A short burst queues and completes. Sustained overload is refused with `503` and a `Retry-After` header, which is a better answer than accepting work the service cannot get to in time.

---

## Server Timeouts

All values are milliseconds. `0` disables a timeout, which is what the Go standard library does by default and the reason a slow client could previously hold a connection open indefinitely.

| Setting | Default | Purpose |
|---|---|---|
| `DIMS_READ_HEADER_TIMEOUT` | `5000` | Time allowed to send request headers |
| `DIMS_READ_TIMEOUT` | `15000` | Time allowed to send the whole request |
| `DIMS_WRITE_TIMEOUT` | `60000` | Time allowed to write the response |
| `DIMS_IDLE_TIMEOUT` | `120000` | Time a keep-alive connection may sit idle |
| `DIMS_SHUTDOWN_TIMEOUT` | `30000` | Time in-flight requests get to finish on shutdown |
| `DIMS_MAX_HEADER_BYTES` | `65536` | Largest request header accepted |

Raise `DIMS_WRITE_TIMEOUT` if you serve very large images, since it has to cover the download, the transformation, and the write.

On `SIGTERM` or `SIGINT` the service stops accepting connections and lets requests already in flight finish, up to `DIMS_SHUTDOWN_TIMEOUT`. Each in-flight request holds libvips buffers, so dropping one mid-encode wastes the work and truncates the response.

---

## `DIMS_DEFAULT_OUTPUT_FORMAT`

Specifies the default image format to convert to when no format is explicitly requested.

- **Default:** *(unset)*

Example:
```
DIMS_DEFAULT_OUTPUT_FORMAT=webp
```

This is useful if you want all images to be served in a modern format by default. If a format is explicitly requested via the URL (e.g. using a `format` command), it takes precedence.

---

## `DIMS_EXCLUDED_OUTPUT_FORMATS`

Comma-separated list of **source** image formats that `DIMS_DEFAULT_OUTPUT_FORMAT` does not apply to.

- **Default:** *(unset)*

Example:
```
DIMS_DEFAULT_OUTPUT_FORMAT=webp
DIMS_EXCLUDED_OUTPUT_FORMATS=gif,svg
```

That converts everything to WebP except images that arrived as GIF or SVG, which keep their normal handling. Matching ignores case.

The list names the format an image **arrives** in, not the format it would be converted to. This matches mod_dims, which reads the input format before deciding whether to apply its default.

An explicit [`format`](../operations/output/format.md) command in the request always wins, whatever this is set to.