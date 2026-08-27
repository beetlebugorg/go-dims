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
/v5/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg&sig=6d3dcb6
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

## ♻️ Conditional Requests

The response carries an `ETag` derived from the commands, the image URL, and the origin's own `ETag`.

A request that sends `If-None-Match` with a matching value gets `304 Not Modified` and no body. `*` matches anything, a comma separated list is searched, and the comparison is weak, so a `W/` prefix on either side is ignored.

The source image is still fetched, because the `ETag` cannot be computed without the origin's. What a `304` skips is the decode, the transformation, and the encode, which is where the time is spent.

If the origin sends no `ETag`, no `ETag` is sent and every request is answered in full.

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

## 🔐 Signing

All `/v5/dims` requests must be signed to ensure the request has not been tampered with.

### How Signing Works

The signature is a **HMAC-SHA256 hash (32 bytes)**. The signing key is the HMAC key. The message holds three fields, one per line, joined with a newline:

1. The **command path**, as it appears after `/v5/`
2. The **raw image URL**, not URL-encoded
3. The **canonical query**: every signed parameter written as `name=value`, percent-encoded and ordered by name

The canonical query is what a standard query string encoder produces from the signed parameters, sorted by name. A parameter carrying several values contributes each of them, in the order the URL gives them.

The parameter name is part of the message, so `a=ab&b=c` and `a=a&b=bc` sign differently. Percent-encoding keeps a value from carrying a separator of its own.

The command path and the image URL must not hold a control character. A request that carries one is refused with `400`.

Every query parameter is signed apart from the exclusions below. You do not list them anywhere. The `_keys` parameter is accepted for backward compatibility and no longer selects what is signed.

### 🧾 Signed Query Parameters

Included in the signature:
- Every query parameter **except** the following:
    - `sig` (the signature itself)
    - `url` (the image URL)
    - `eurl` (an encrypted version of `url`, not used in signing)
    - `_keys` (retained for backward compatibility, ignored when signing)
    - `download` (controls content disposition, excluded from signing)

This means `overlay` is signed. A signed watermark URL cannot be replayed against a different overlay image.

Example:
```
/v5/resize/100x100/?url=https://example.com/image.jpg&overlay=http://example.com/overlay.png
```

The HMAC key is `DIMS_SIGNING_KEY`. The signed message becomes:

```
resize/100x100/
https://example.com/image.jpg
overlay=http%3A%2F%2Fexample.com%2Foverlay.png
```

### Building the message

Each language builds the canonical query with its own query string encoder. These all produce the same signature:

```php
$p = ["b" => "2", "a" => "1"];
ksort($p);
$message = implode("\n", [$commands, $imageUrl, http_build_query($p)]);
$sig = hash_hmac("sha256", $message, $key);
```

```python
message = "\n".join([commands, image_url, urlencode(sorted(params.items()))])
sig = hmac.new(key, message.encode(), hashlib.sha256).hexdigest()
```

```javascript
const q = new URLSearchParams(params);
q.sort();
const message = [commands, imageUrl, q.toString()].join("\n");
const sig = createHmac("sha256", key).update(message).digest("hex");
```

```ruby
message = [commands, image_url, URI.encode_www_form(params.sort)].join("\n")
sig = OpenSSL::HMAC.hexdigest("SHA256", key, message)
```

## 🔐 `eurl` encryption

The `eurl` query parameter allows you to encrypt a full image URL, so that it is not exposed in
plaintext. This is useful when you want to:

- Obscure or protect source URLs (e.g. signed S3 links)
- Watermark images with a URL that should not be visible to users

### Implementing `eurl` Encryption

To generate an `eurl` compatible with go-dims, follow these steps:

1. **Key Derivation**  
  - Use the HKDF-SHA256 key derivation function to derive a 16-byte (128-bit) AES key from the secret shared in `DIMS_SIGNING_KEY`.
  - Use the string `go-dims` for the salt.

2. **Encryption**  
  - Use AES-128-GCM to encrypt the original image URL.
  - Generate a 12-byte random IV (nonce).
  - Encrypt the URL using the derived key and IV.

3. **Output Format**  
   - Concatenate the IV, ciphertext, and tag in that order.
   - Base64-encode the entire byte sequence. 
   - The resulting string should be used as the value for the `eurl` parameter.

Any mismatch in the key, salt, IV size, or output format will result in a decryption failure (`cipher: message authentication failed`).

### ✅ Use the CLI

To simplify signing, you can use the `sign` command. It will compute the signature correctly based
on the same rules used by the server:

The signature depends on the key, so this example shows the key that produced it.

``` 
❯ cat dims.key
0123456789abcdef0123456789abcdef

❯ ./dims sign --key-file=dims.key 'https://myhost.com/v5/resize/100x100/?url=https://example.com/image.jpg&overlay=http://example.com/overlay.png'

https://myhost.com/v5/resize/100x100/?overlay=http%3A%2F%2Fexample.com%2Foverlay.png&sig=45bda414d496b433fef6d38f4844fc6e079709ae0e91e3d65355163f26186633&url=https%3A%2F%2Fexample.com%2Fimage.jpg
```