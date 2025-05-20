# Strip

| Command   | Argument Type 
|------------|--------------
| `strip`    | `<bool>`

This command will remove ICC, and Exif profiles.

Use this command to reduce the size of your images. This especially helps
when generating thumbnail images as this profile data can become a large
percent of the final image once it's been reduced to a thumbnail.

By default dims will strip metadata anyway so this call is a no-op.

However, there are use cases such as DAM solutions where you may not want to
strip this data by default. In those cases disable [DIMS_STRIP_METADATA](../configuration/general.md#dims_strip_metadata)
globally and call this command when needed.