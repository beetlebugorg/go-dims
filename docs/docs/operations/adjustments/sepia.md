# Sepia

Apply a warm brown monochrome tone.

## Syntax

| Command | Argument Format |
|---------|-----------------|
| `sepia` | `<0.0-1.0>`     |

## Behavior

- The argument is the strength of the effect. `0` leaves the image unchanged and `1` applies the full tone.
- Values between the two blend the original with the toned result.
- An argument outside 0 to 1 returns `400`.
- Alpha is preserved.

## Examples

```
/v5/sepia/1/?url=photo.jpg
/v5/thumbnail/200x200/sepia/0.6/?url=photo.jpg
```

## Difference from mod_dims

`mod_dims` passed this argument to ImageMagick as a tone threshold. `go-dims` does not use ImageMagick, and renders the effect with its own colour matrix, so the output differs.

The argument keeps the same range, and a larger value still means a stronger effect, so an existing URL stays valid and still produces a sepia image. It will not be the same image.
