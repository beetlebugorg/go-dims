---
sidebar_position: 2
---

# Crop

Crop an image to a specific region using width, height, and optional x/y offsets.

## Syntax

| Command | Argument Format               |
|---------|-------------------------------|
| `crop`  | `<width>x<height>[+x+y]`      |

- `<width>` and `<height>` define the size of the crop region.
- `+x+y` is optional and specifies the offset from the top-left corner.
    - Defaults to `+0+0` if omitted.

## Behavior

- Crops the image to the specified size starting at the given offset.
- The resulting region must stay within the bounds of the original image:
  - `width + x` must not exceed image width
  - `height + y` must not exceed image height
- Offsets can be **absolute** pixel values or **percentages** of the image dimensions.
  - Percentage values must be URL-escaped, e.g. `%` → `%25`

## Examples

#### Crop a 256x100 region, offset by 0 pixels horizontally and 120 pixels vertically:

![Listing 1 - resize/256x256^/crop/256x100+0+120](../../assets/crop256x100+0+120.jpg "Listing 1")
```
/v5/resize/256x256^/crop/256x100+0+120/?url=pexels-photo-1539116.jpeg
```

#### Center crop a 512×256 region (using a vertical offset of 25%):

![Listing 2 - crop/512x256+0+25%](../../assets/crop512x256+0+25.jpg "Listing 2")
```
/v5/crop/512x256+0+25%25/?url=pexels-photo-1539116.jpeg
```