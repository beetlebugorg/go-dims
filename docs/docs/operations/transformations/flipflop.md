---
sidebar_position: 4
---

# Flip Flop

Flip an image horizontally, vertically, or both.

## Syntax

| Command    | Argument Format     |
|------------|----------------------|
| `flipflop` | `horizontal` or `vertical` |

## Behavior

- Flips the image along the specified axis:
    - `horizontal` — mirrors the image left-to-right.
    - `vertical` — mirrors the image top-to-bottom.
- You may chain the operation to apply both (e.g., `flipflop/horizontal/flipflop/vertical`).

## Examples

#### Horizontal Flip

```
/v5/flipflop/horizontal/?url=pexels-photo-1539116.jpeg
```

![Horizontal Flip](../../assets/flip-horizontal.jpg)

Image mirrored horizontally (left-to-right).

#### Vertical Flip

```
/v5/flipflop/vertical/?url=pexels-photo-1539116.jpeg
```

![Vertical Flip](../../assets/flip-vertical.jpg)

Image mirrored vertically (top-to-bottom).

#### Flip Both Directions

```
/v5/flipflop/horizontal/flipflop/vertical/?url=pexels-photo-1539116.jpeg
```

![Both Flips](../../assets/flip-both.jpg)

Image mirrored horizontally and vertically.