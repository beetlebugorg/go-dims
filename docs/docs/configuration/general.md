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

## `DIMS_OUTPUT_FORMAT_EXCLUDE`

Specifies a comma-separated list of image formats that should **not** be converted by the default output format setting.

- **Default:** *(unset)*

Example:
```
DIMS_OUTPUT_FORMAT_EXCLUDE=GIF,SVG
```

This allows certain image types (like animated GIFs or vector SVGs) to bypass the default format conversion logic.