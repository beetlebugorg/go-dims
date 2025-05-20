# Crop

| Command        | Argument Type 
|----------------|--------------
| `crop`         | `<width>x<height>+{x}+{y}`

This command will crop an image to the size, and at the location, specified in the image geometry.

The crop image geometry format is `<width>x<height>+{x}+{y}` which will crop an image to the
size specified `width` and `height` offset by `x` and `y`.

The `width` plus `x` offset must be less than the width of the image.

The `height` plus `y` offset must be less than the height of the image.

The `+{x}+{y}` is optional, they will both default to `0`.