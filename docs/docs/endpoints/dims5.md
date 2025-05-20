# `/v5`

```
/v5/<commands>.../?url=<image>&sig=<signature>
```

This endpoint allows you to apply image transformations such as cropping, resizing, format conversion, and more.

Each request includes:
- One or more transformation **commands**, specified in the path
- A target image provided via the `url` query parameter
- A **signature** that validates the request

Commands are applied **in order**, from left to right.

---

## 🧩 Example Breakdown

From the [Getting Started](../installation.md) guide:

```
/v5/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg&sign=6d3dcb6&expire=2147483647
```

### Path: Commands

| Parameter   | Value             | Description                                              |
|-------------|-------------------|----------------------------------------------------------|
| `commands`  | `resize/100x100`  | One or more [operations](../operations), slash-separated |

### Query String

| Parameter  | Value                           | Description                                                                       |
|------------|---------------------------------|-----------------------------------------------------------------------------------|
| `url`      | `https://images.pexels.com/...` | The image to manipulate                                                           |
| `sig`      | `6d3dcb6...`                    | [Signature](../configuration/signing) to verify the request                       |
| `download` | `1` (optional)                  | Forces the image to download instead of displaying inline (`Content-Disposition`) |

---

## 🛑 Error Handling

This endpoint will **always try to return an image**, even when something goes wrong.

If an error occurs (e.g., download failure, invalid command), a fallback image is generated using
the background color defined by
[`DIMS_ERROR_BACKGROUND`](../configuration/general.md#dims_error_background).

The error image will be cropped/resized as needed to match the requested dimensions, so layout
remains consistent on your page. In some cases it may not be able to match the requested dimensions,
for example when a transformation command's argument has a syntax error. In those cases a 512x512
image will be returned.

---

## 🔐 Signing

All `/v5/dims` requests must be signed to ensure the request has not been tampered with.

### How Signing Works

The signature is a **HMAC-SHA256 hash (32 bytes)** of the following, concatenated in order:

1. The **signing key**
2. The **command path** (no leading or trailing slashes)
3. The **raw image URL** (not URL-encoded)
4. The **values of any signed query parameters**, in iteration order

---

### 🧾 Signed Query Parameters

Only a specific set of query parameters are included in the signature:

Included in signature:
- Any query parameter **except** the following:
    - `sig` (the signature itself)
    - `url` (the image URL)
    - `eurl` (an encoded version of `url`, not used in signing)
    - `_keys` (optional debugging/tracking param)
    - `download` (controls content disposition, excluded from signing)

Example:
```
/v5/resize/100x100/?url=https://example.com/image.jpg&overlay=http://example.com/overlay.png
```

Signature input becomes:
- `<secret>` (value of `DIMS_SIGNING_KEY`)
- `resize/100x100`
- `https://example.com/image.jpg`
- `http://example.com/overlay.png` (value of `overlay`)

---

### ✅ Use the CLI

To simplify signing, you can use the `sign` command. It will compute the signature correctly based
on the same rules used by the server:

``` 
❯ ./dims sign --key-file=dims.key 'https://myhost.com/v5/resize/100x100/?url=https://example.com/image.jpg&overlay=http://example.com/overlay.png'

https://myhost.com/v5/resize/100x100/?overlay=http%3A%2F%2Fexample.com%2Foverlay.png&sig=f598fe37ff0e9e0a5794504f779f76ca0ce5596518b65850900d2c3247e12dce&url=https%3A%2F%2Fexample.com%2Fimage.jpg
```