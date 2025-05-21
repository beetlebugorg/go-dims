---
sidebar_position: 4
---

# Image Sources

`go-dims` supports multiple source backends for fetching original images: `http`, `s3`, and `file`.

You can explicitly specify the backend using a scheme-style prefix:

- `http://example.com/image.jpg`
- `s3://bucket/key/image.jpg`
- `file://path/to/image.jpg`

If the URL **does not** include a scheme prefix (e.g., just `?url=cat.jpg`), `go-dims` will fall back to the backend defined by:

### `DIMS_DEFAULT_SOURCE_BACKEND`

- **Default:** `http`

To support simplified URLs with S3 or file sources, you must set that backend as the default:

```
DIMS_DEFAULT_SOURCE_BACKEND=s3
```

This allows you to make requests like:

```
?url=my-folder/image.jpg
```

And `go-dims` will resolve it using the configured S3.

---

### `DIMS_ALLOWED_SOURCE_BACKENDS`

Comma-separated list of backends that are permitted.

- **Default:** `http`

Only the specified backends will be allowed to handle requests. If a backend is not listed, requests using that scheme will return an error.

:::note

By default, s3 and file are disabled. You must explicitly enable them to allow access.


```
DIMS_ALLOWED_SOURCE_BACKENDS=http,s3
```

To allow all supported backends:

```
DIMS_ALLOWED_SOURCE_BACKENDS=http,s3,file
```

:::

---

## S3 Source Configuration

Enable fetching images from Amazon S3 by configuring the following variables:

### `DIMS_S3_BUCKET`

The S3 bucket name from which to fetch images.

- **Default:** *(empty)*

---

### `DIMS_S3_PREFIX`

Optional prefix to apply to all S3 object keys.

- **Default:** *(empty)*

This is useful if you store images in a specific folder or namespace within your bucket.

Example:
```
DIMS_S3_BUCKET=my-bucket  
DIMS_S3_PREFIX=images/2024/
```

A request for `image.jpg` would resolve to `s3://my-bucket/images/2024/image.jpg`.

---

## File Source Configuration

Enable local file access (useful for development or staging).

### `DIMS_FILE_BASE_DIR`

Specifies the base directory for reading local files.

- **Default:** `./resources`

Example:
```
DIMS_FILE_BASE_DIR=/var/images
```

A request for `sample.jpg` would resolve to `/var/images/sample.jpg`.
