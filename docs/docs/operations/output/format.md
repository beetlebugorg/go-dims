# Format

Convert the output image to a different format.

## Syntax

| Command  | Argument Format |
|----------|------------------|
| `format` | `jpg`, `png`, or `webp` |

## Behavior

- Forces the output image to be encoded in the specified format.
- Useful for converting images on-the-fly or standardizing delivery formats.
- If no format is specified, the output format defaults to the value of the `DIMS_DEFAULT_OUTPUT_FORMAT` environment variable.

### Supported Formats

- `jpg` — lossy, no alpha support, widely compatible
- `png` — lossless, supports transparency
- `webp` — modern format, supports both lossy/lossless and alpha