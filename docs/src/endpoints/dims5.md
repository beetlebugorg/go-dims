# /v5/dims Endpoint

```html
/v5/dims/<commands>.../?url=<image>&clientId=<clientId>&sig=<signature>
```

The `dims` endpoint lets you crop, resize, and apply other transformations to images.

An image manipulation request is made up of one or more [commands](#commands) that will transform an image, such
as `resize/100x100`. These commands will be applied on the image provided in the
`url` parameter. They are applied in the order they appear in the URL.

Let's break down the example we used in the [Getting Started](../guide/installation.md) section:

```html
/v5/dims/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg&sign=6d3dcb6&expire=2147483647&clientId=default
```

Breaking the request down into its parts we get the following:

**URL**:
| Parameter               |  Value            | Description
|-------------------------|-------------------|----------------------------------------------------------------
| `commands`              | **`resize/100x100`**    | one or more [commands](#commands), separated by `/`

**Query String**:
| Parameter         |  Value            | Description
|-------------------|-------------------|----------------------------------------------------------------
| `url`             | `https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg` | image to manipulate
| `download`        | `0`                | if set to `1` include `attachment` content disposition header
| `clientId`        | **`default`**           | name of client making request, tied to [signing key](../configuration/signing.md)
| `sig`            | **`6d3dcb6...`**           | [signature](#signing) to prevent request tampering

## Error Handling

This endpoint will always return an image.  When a command fails dims will return an auto-generated image
using the background color defined in the
[DIMS_ERROR_BACKGROUND](../configuration/other.md#dims_error_background) environment variable.

The auto-generated error image will be resized and/or cropped to match the requested image so it'll
fit nicely in the space where the original image would have been.

## Signing

All requests to this endpoint must be signed. Signing requests ensures that the image request has
not been changed.

The signature is a HMAC-SHA256-128 of:
- `imageCommands`
- `imageUrl`
 
Those values should be concatenated together without any spaces or other characters between them, and
then signed.

Note:
- `imageCommands` **should not** have any preceding or trailing slashes (`/`).
    - ✅️`resize/100x100/crop/10x10+25+25`

    - ❌️`/resize/100x100/crop/10x10+25+25/`

- `imageUrl` **should not** be url encoded.
    - ✅️`https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg`

    - ❌️`https%3A%2F%2Fimages.pexels.com%2Fphotos%2F1539116%2Fpexels-photo-1539116.jpeg`

We've wrapped this up into a command to make signing and validating requests easier.

Use the `sign` command to sign a request by passing in a URL to sign, and setting the signing
key in the `DIMS_SIGNING_KEY` environment variable.

```shell
❯ docker run -e DIMS_SIGNING_KEY=mysecret ghcr.io/beetlebugorg/go-dims sign 'https://myhost.com/v5/dims/resize/100x100/?clientId=default&url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg' --dev  

Image to be transformed:

https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg

Transformation commands found:

resize('100x100')

http://localhost:8080/v5/dims/resize/100x100/?clientId=default&url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg&sig=325d12cdbcb496ac7c27622b7bb6fb453449b3789dc3fa999996691f18f2dc
```

## Commands

{{#include ../operations/v5.md}}