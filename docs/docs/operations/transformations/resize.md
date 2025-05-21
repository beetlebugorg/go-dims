---
sidebar_position: 1
---

# Resize

Resize an image by specifying a target geometry. By default, aspect ratio is preserved.

## Syntax

| Command  | Argument Format       |
|----------|------------------------|
| `resize` | `<width>x<height>[!]`  |

- `<width>` and `<height>` are integers.
- `!` (optional) forces exact dimensions, disabling aspect ratio preservation.
- `^` (optional) similar to `!`, but may output an image larger than the requested size. Use this
                 in combination with `crop` to ensure the image is cropped to the requested size.

## Behavior

- **Without `!`** — image is resized to fit within the specified box, preserving its aspect ratio.
The output may be smaller than requested in one dimension.  
 
- **With `!`** — image is resized to
exactly match the specified dimensions, which may result in stretching or squishing.

## Examples

#### Forced Resize (with `!`)

![Listing 1 - resize/100x100!](../../assets/resize100x100exclamation.jpg "Listing 1")

#### Proportional Resize (no `!`)

![Listing 2 - resize/100x100](../../assets/resize100x100.jpg "Listing 2")

:::warning URL Escaping

Some geometry formats include symbols like `%` that must be URL-encoded.

- For example, to resize to **15%** of the original size:  
  Use `resize/15%25`, where `%25` is the encoded form of `%`.

Always ensure resize arguments are properly escaped when constructing URLs.

:::