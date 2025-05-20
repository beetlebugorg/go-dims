# Watermark

Overlay a secondary image (e.g. logo or badge) on top of the source image.

## Syntax

| Command     | Argument Format               |
|-------------|-------------------------------|
| `watermark` | `<opacity>,<size>,<gravity>`  |

This command requires a query parameter:

- `overlay` — the URL of the image to use as the watermark

## Behavior

This operation applies an image watermark by:

1. Fetching the overlay image from the `overlay` query parameter.
2. Scaling the overlay relative to the base image using the given size ratio.
3. Reducing the overlay's opacity.
4. Positioning it based on the specified gravity.
5. Compositing it over the original image.

## Parameters

| Parameter   | Type     | Range     | Description                                                                 |
|-------------|----------|-----------|-----------------------------------------------------------------------------|
| `opacity`   | `float`  | 0.0–1.0   | Controls watermark transparency — `0.0` is fully transparent, `1.0` is solid. |
| `size`      | `float`  | 0.0–1.0   | Scales the watermark relative to the base image’s largest dimension.        |
| `gravity`   | `string` |           | Position of the watermark. One of: `n`, `ne`, `nw`, `s`, `se`, `sw`, `w`, `e`, `c`. |

## Example

#### Overlay a watermark scaled to 25% of the base image’s size, positioned at the southeast corner:

```
/v5/watermark/1,.25,se?url=pexels-photo-1539116.jpeg&overlay=hex-lab.svg
```

![Watermark](../../assets/watermark.jpg)

## Secure URL Encryption with `eurl`

To prevent exposing the original image URL (e.g. when applying watermarks to private or internal images), you can use the `eurl` query parameter instead of `url`.

The `eurl` value is an **AES-128-GCM–encrypted and base64-encoded** string that securely represents the original image URL.

### Behavior

- go-dims will decrypt the `eurl` value using the `DIMS_SIGNING_KEY`
- The decrypted URL is treated exactly as if it had been passed via `url`
- This helps obfuscate origin URLs while still allowing signed and cacheable requests

### Encryption Details

- Uses **AES-128-GCM**.
- The key is derived from your configured `DIMS_SIGNING_KEY` using the following process:
    - Hash the signing key using **SHA-1**
     Convert the hash to a hex string
    - Take the **first 16 characters** of the hex string
    - Convert to uppercase and use as the 16-byte AES key

### Signature Calculation

> When using signed requests, the **unencrypted image URL** must be used during signature generation — not the `eurl` value.

This ensures consistency on the server side, where decryption happens **before** signature validation.

### Example

```
/v5/watermark/0.3,0.2,se/?eurl=BASE64_ENCRYPTED_URL&overlay=logo.png&sig=...
```

This protects the original image URL from being visible in client code, logs, CDNs, or analytics tools.

