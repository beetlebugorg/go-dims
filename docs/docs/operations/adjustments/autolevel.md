---
sidebar_position: 1
---

# Autolevel

Automatically adjusts the contrast of the image by stretching its histogram to span the full dynamic range.

## Syntax

| Command    | Argument Format |
|------------|-----------------|
| `autolevel` | true or false  |

## Behavior

- Enhances image contrast by remapping the darkest and lightest pixels to pure black and white.
- Equivalent to applying a basic “auto contrast” filter.
- Useful for correcting low-contrast or washed-out images.

## Example

Apply autolevel to a flat or grayish image:

```
/v5/autolevel/true?url=https://images.pexels.com/photos/1835899/pexels-photo-1835899.jpeg
```

![Autolevel](../../assets/autolevel-true.jpg)

Image with contrast automatically enhanced, compare with the original:

![Autolevel](../../assets/autolevel-false.jpg)
