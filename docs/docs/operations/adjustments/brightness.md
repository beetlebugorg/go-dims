---
sidebar_position: 3
---

# Brightness

Adjust the brightness and contrast of the image using a polynomial tone curve.

## Syntax

| Command     | Argument Format           |
|-------------|----------------------------|
| `brightness` | `<brightness>x<contrast>`  |

## Behavior

This command applies a tone curve.

- `<brightness>` controls the image's midpoint intensity.
  - Value is **-100 and 100**. 0 is no change, 100 is maximum brightness (effectively all white image).
- `<contrast>` controls the slope of the curve (i.e. how stretched the values are).
  - Value is **-100 and 100**.

## Common Examples

#### Brighten at 5, contrast at 20:

```
/v5/brightness/5x20/?url=pexels-photo-1539116.jpeg
```

Original on the left, brightened and contrast increased on the right:

![Original](../../assets/original128x.jpg)
![Brightness +20, Contrast +10](../../assets/brightness5x20.jpg)

#### Brightness decreased, contrast increased:

```
/v5/brightness/-5x5/?url=pexels-photo-1539116.jpeg
```

![Original](../../assets/original128x.jpg)
![Dark + Low Contrast](../../assets/brightness-5x5.jpg)

## Notes

- If you only care about brightness, you can use `brightness/60x0`