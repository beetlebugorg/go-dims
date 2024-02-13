# go-dims

Go port of mod-dims.

This port:

    - runs as a standalone service
    - does not require Apache httpd
    - does not implement local filesystem access
    - does not honor timestamp expiration
    - aims to be backward compatible with the `/dims4/` version of mod-dims
    - uses Imagemagick 7.x

## Running

To run in development mode that disables signature verification, the default is production mode:

```
docker run -p 8080:8080 ghcr.io/beetlebugorg/go-dims serve --dev --debug --bind ":8080"
```

This software uses the following software:

    *imagick*

    Copyright (c) 2013-2014, The GoGraphics Team
    All rights reserved.