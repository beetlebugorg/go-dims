---
sidebar_position: 5
---

# Rotate

Rotate an image by a specified number of degrees.

## Syntax

| Command  | Argument Format  |
|----------|------------------|
| `rotate` | `float` (degrees) |

## Behavior

- Rotates the image **clockwise** by default.
- Use **negative values** to rotate counter-clockwise.
- Accepts any float between `0` and `360`.  
  Example: `rotate/25.75` is valid.

If the rotation angle is **not a multiple of 90**, the resulting image will be enlarged to accommodate the rotated bounds.  
Any exposed corners will be filled with a **transparent** background (for formats that support alpha) or a solid background color.

Note: Subtle angles are supported, but changes below 1 degree are rarely noticeable in practice.

## Examples

#### Rotate the image 90 degrees clockwise:

```
/v5/rotate/90/?url=pexels-photo-1539116.jpeg
```

![Rotate 90](../../assets/rotate90.jpg)

#### Rotate 45 degrees counter-clockwise:


```
/v5/rotate/-45/?url=pexels-photo-1539116.jpeg
```

![Rotate -45](../../assets/rotate45.jpg)

Note the filled corners outside the original bounds, and the enlarged canvas to fit the rotated image.
