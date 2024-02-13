# go-dims

`go-dims` is a Go implementation of the DIMS API as implemented by
[mod-dims](https://github.com/beetlebugorg/mod_dims).

The DIMS API traces its history back to Aol. Prior to `mod-dims` there was a
legacy implementation written in Java using jmagick. You can find bits of that
original URL protocol in the [mod-dims codebase](https://github.com/beetlebugorg/mod_dims/blob/master/src/mod_dims.c#L1658-L1668).

This implementation differs from `mod-dims` in the following ways:

    - runs as a standalone service
    - can be embedded in other Go applications
    - does not require or use Apache httpd
    - does not implement local filesystem access
    - does not implement the `/dims` or `/dims3` endpoints
    - does not honor timestamp expiration
    - aims to be backward compatible with the `/dims4/` endpoint of mod-dims
    - uses Imagemagick 7.x

## Running

To run in development mode that disables signature verification, the default is production mode:

```
docker run -p 8080:8080 ghcr.io/beetlebugorg/go-dims serve --dev --debug --bind ":8080"
```

## Configuration

The following environment variables are used to configure the service:

### Common Settings

`DIMS_SECRET_KEY` is the secret key for signing URLs. This is required.

`DIMS_DOWNLOAD_TIMEOUT` is the maximum time in milliseconds to wait for a download.

`DIMS_PLACEHOLDER_BACKGROUND` is the background color for placeholder images. 
This is used when there is an error processing an image. The default is `#5ADAFD`.

### Miscellaneous

`DIMS_INCLUDE_DISPOSITION` - whether to include the Content-Disposition header

`DIMS_STRIP_METADATA` - whether to strip metadata from images

`DIMS_DEFAULT_OUTPUT_FORMAT` - the default output format for images

`DIMS_IGNORE_DEFAULT_OUTPUT_FORMATS` - a comma-separated list of output formats to ignore

`DIMS_DEFAULT_IMAGE_PREFIX` - the default image prefix

### Cache Control

The following environment variables are used to configure the cache control headers:

`DIMS_CACHE_CONTROL_MAX_AGE` is the maximum age in seconds for the Cache-Control header. 
The default is 31536000.

`DIMS_PLACEHOLDER_IMAGE_EXPIRE` is the maximum age in seconds for placeholder images. The default is 60.

These are used to set the `max-age` in the `Cache-Control` header.

```
Cache-Control: max-age=${DIMS_DEFAULT_EXPIRE}, public
```

`DIMS_EDGE_CONTROL_DOWNSTREAM_TTL` is the `max-age` for the Edge-Control header.

```
Edge-Control: max-age=${DIMS_EDGE_CONTROL_DOWNSTREAM_TTL}
```

`DIMS_TRUST_SRC` is whether to trust the Cache-Control header of the source
image.  The default is `false`. If the source is trusted, has a `Cache-Control`
header, and its `max-age` is between `DIMS_MIN_SRC_CACHE_CONTROL` and
`DIMS_MAX_SRC_CACHE_CONTROL`, then the `max-age` of the source image is used.

`DIMS_MIN_SRC_CACHE_CONTROL` is the minimum Cache-Control header for the source image.

`DIMS_MAX_SRC_CACHE_CONTROL` is the maximum Cache-Control header for the source image.

`DIMS_SIGNING_ALGORITHM` is the signing algorithm to use. The default is `hmac-sha256`. 
Set this to `md5` to be backward compatible with mod-dims.

### Other

This software uses the following software:

    *imagick*

    Copyright (c) 2013-2014, The GoGraphics Team
    All rights reserved.