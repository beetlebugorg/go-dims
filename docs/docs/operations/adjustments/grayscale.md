---
sidebar_position: 4
---

# Grayscale

Convert the image to grayscale by removing all color information.

## Syntax

| Command    | Argument Format |
|------------|-----------------|
| `grayscale` | true or false  |

## Behavior

- Removes color by converting the image to a single luminance-based channel.
- The result is a neutral grayscale image with shades from black to white.
- Useful for stylistic adjustments, reducing file size, or image processing tasks that don’t require color.

## Example

#### Convert an image to grayscale:

```
/v5/grayscale/true?url=pexels-photo-1539116.jpeg
```

![Grayscale Example](../../assets/grayscale.jpg)  

Image rendered in grayscale