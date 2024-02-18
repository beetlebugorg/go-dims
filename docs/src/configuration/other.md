# Other Settings

Occasionally useful settings that can probably be left as-is, except for the
error color setting, change that one.

### DIMS_ERROR_BACKGROUND

This sets the background color of the error image.

The default is `#5adafd`:

<span style="background:#5adafd; display: block; width: 50px; height: 50px"></span>

The error image may be sent because:
- dims was unable to download the source image
- imagemagick was unable to resize the image
- memory or resource pressure

Think of it as the error color for dims. You don't want to see it but when you do
you probably want to check your logs to find out why.

If you're looking for a way to set a error image, that doesn't exist anymore. It 
was called the `NOIMAGE` image in mod-dims. The error image is now generated
automatically using Imagemagick's PixelWand.

### DIMS_STRIP_METADATA

This enables, or disables metadata removal after image manipulation.

The default is `true`.

Enabling this will remove all profile and Exif data from all images served by
dims. This is usually what you want but if you're using dims for a 
DAM (Digital Asset Management) solution you may want to control which images
you remove data from.

If you disable this you can use the [strip](../operations/strip.md) operation to remove profiles per image.

### DIMS_INCLUDE_DISPOSITION

This enables the `Content-Disposition` header.

The default is `false`.

When you enable this the `Content-Disposition` header like below will be included
in every response:

```
Content-Disposition: inline; filename=<filename>
```

If you want to trigger an image to be downloaded you need that to be `attachment; filename=<filename>`.
You can get that by sending the `download` argument in your request (`?url=...&download=1`).

### DIMS_DOWNLOAD_TIMEOUT

The maximum time an origin image download is allowed before it is cancelled.

The default is `3000`. This is in milliseconds.

### DIMS_DEFAULT_OUTPUT_FORMAT

The default image format to convert images to if one is not provided.

This is not set by default.

You may find that you want to force all images to be a certain type such as `webp`. This
setting can help but keep in mind that if the image request also includes
a `format` command that will take precedence.

### DIMS_OUTPUT_FORMAT_EXCLUDE

Image formats that will not be converted to the default output format.

This is not set by default. 

Provide as a comma delimited list: `GIF,SVG`