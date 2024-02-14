# /v4 Endpoint

```html
/v4/<ClientId>/<Signature>/<Timestamp>/<Commands>.../?url=<Image>
```

This endpoint is backward compatible with the mod-dims `/dims4` endpoint. You can also
access this endpoint at `/dims4`.

## Image Manipulation Request

An image manipulation request is made up of the url to the image you want to manipulate, and one or
more commands.

Here is the example we used in the [Getting Started](../guide/installation.md) section:

```html
/v4/default/6d3dcb6/2147483647/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg
```

Breaking the request down into its parts you get the following:

**URL parameters**:
| Parameter               |  Value            | Description
|-------------------------|-------------------|----------------------------------------------------------------
| `clientId`              | **`default`**           | name of client making request, tied to [signing key](../configuration/signing.md)
| `signature`             | **`6d3dcb6`**           | [signature](#signing) to prevent request tampering
| `timestamp`             | **`2147483647`**        | expiration as a unix timestamp (seconds since epoch)
| `commands`              | **`resize/100x100`**    | one or more [commands](#commands), separated commands with `/`

**Query parameters**:
| Parameter             |  Value            | Description
|-----------------------|-------------------|----------------------------------------------------------------
| `url`                | `https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg` | image to manipulate
| `download`           | `0`                | if set to `1` include `attachment` content disposition header

## Commands

{{#include ../operations/operations.md}}

## Signing

All requests to this endpoint must be signed. Signing requests ensures that
the image request has not been changed.

In the examples below note the following:
- *imageCommands* **should** have any preceding or trailing slashes (`/`) removed. 
- *imageUrl* **should not** be url encoded.
- *timestamp* is a unix timestamp (seconds since epoch).

The algorithm used is controlled by the environment variables `DIMS_SIGNING_ALGORITHM`. See
[Signing Request](../configuration/signing.md#dims_signing_algorithm) configuration.

### HMAC-SHA256

Requests are signed using **HMAC-SHA256**.

To sign a request concatenate and sign: *timestamp*, *imageCommands*, and *imageUrl*.

Here is how we do this on the server in Go:

```go
func SignHmacSha256(timestamp string, secret string, imageCommands string, imageUrl string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(imageCommands))
	mac.Write([]byte(imageUrl))

	return fmt.Sprintf("%x", mac.Sum(nil))[0:24]
}
```

To test this out you can use the `sign` subcommand of dims:

```shell
$ dims sign 2147483647 mysecret resize/100x100 "https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg"
dc45b8f4a9405247f25bd22f
```

### MD5

Legacy mod-dims requests were signed with md5.

To sign a request with md5 concatenate and hash: *timestamp*, *secret*, *imageCommands*, *imageUrl*.

Here is how we do this on the server in Go:

```go
func Sign(timestamp string, secret string, imageCommands string, imageUrl string) string {
	hash := md5.New()
	io.WriteString(hash, timestamp)
	io.WriteString(hash, secret)
	io.WriteString(hash, imageCommands)
	io.WriteString(hash, imageUrl)

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}
```

To test this out you can use the `sign` subcommand of dims:

```shell
$ dims sign 2147483647 mysecret resize/100x100 "https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg" --signing-algorithm=md5
6d3dcb6
```

> This signature method is deprecated and should not be used.
