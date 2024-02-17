# /v4/dims Endpoint

```html
/v4/dims/<clientId>/<signature>/<timestamp>/<commands>.../?url=<image>
```

The `dims` endpoint lets you crop, resize, and apply other transformations to images.

An image manipulation request is made up of one or more [commands](#commands) that will transform an image, such
as `resize/100x100`. These commands will be applied on the image provided in the
`url` parameter. They are applied in the order they appear in the URL.

Let's break down the example we used in the [Getting Started](../guide/installation.md) section:

```html
/v4/dims/default/6d3dcb6/2147483647/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg
```

Breaking the request down into its parts we get the following:

**URL**:
| Parameter               |  Value            | Description
|-------------------------|-------------------|----------------------------------------------------------------
| `clientId`              | **`default`**           | name of client making request, tied to [signing key](../configuration/signing.md)
| `signature`             | **`6d3dcb6`**           | [signature](#signing) to prevent request tampering
| `timestamp`             | **`2147483647`**        | expiration as a unix timestamp (seconds since epoch)
| `commands`              | **`resize/100x100`**    | one or more [commands](#commands), separated by `/`

**Query String**:
| Parameter             |  Value            | Description
|-----------------------|-------------------|----------------------------------------------------------------
| `url`                | `https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg` | image to manipulate
| `download`           | `0`                | if set to `1` include `attachment` content disposition header

## Error Handling

This endpoint will always return an image.  When a command fails dims will return an auto-generated image
using the background color defined in the
[DIMS_ERROR_BACKGROUND](../configuration/other.md#dims_error_background) environment variable.

The auto-generated error image will be resized and/or cropped to match the requested image so it'll
fit nicely in the space where the original image would have been.

## Signing

All requests to this endpoint must be signed. Signing requests ensures that
the image request has not been changed.

The signature is a MD5-56 hash of:
- `timestamp`
- `signingKey`
- `imageCommands`
- `imageUrl`. 
 
Those values should be concatenated together without any spaces or other characters between them,
and then hashed.

Note:
- `imageCommands` **should not** have any preceding or trailing slashes (`/`).
  - ✅️`resize/100x100/crop/10x10+25+25`
  
  - ❌️`/resize/100x100/crop/10x10+25+25/`

- `imageUrl` **should not** be url encoded.
  - ✅️`https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg`

  - ❌️`https%3A%2F%2Fimages.pexels.com%2Fphotos%2F1539116%2Fpexels-photo-1539116.jpeg`

- `timestamp` is a unix timestamp (seconds since epoch).

Bringing that all together, here is an example of how to sign a request using the example from above:

```shell
$ echo -n "2147483647mysecretresize/100x100https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg" | md5 | cut -b 1-7
6d3dcb6
```

We've wrapped this up into a simple command:

```shell
$ dims sign 2147483647 mysecret resize/100x100 "https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg" --signing-algorithm=md5
6d3dcb6
```

## Commands

{{#include ../operations/v4.md}}