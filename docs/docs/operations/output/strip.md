# Strip

Remove metadata from the output image, including ICC profiles and Exif data.

## Syntax

| Command | Argument Format |
|---------|------------------|
| `strip` | `true` or `false` |

## Behavior

- Removes embedded metadata such as:
    - ICC color profiles
    - Exif data (e.g. camera info, GPS)
- Reduces file size — especially useful for thumbnails or web delivery.
- Has no effect on the visual content of the image.

By default, go-dims strips metadata automatically. If metadata stripping is **enabled globally**,
this operation can be used to **opt out** by setting it to `false`.

## Configuration Notes

If you are using go-dims in contexts like Digital Asset Management (DAM) systems and want to
preserve metadata globally:

- Set [`DIMS_STRIP_METADATA`](../../configuration/general.md#dims_strip_metadata) to `false`
- Use `/strip/true` explicitly in requests where metadata removal is desired