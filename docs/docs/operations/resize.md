# Resize

| Command        | Argument Type 
|----------------|--------------
| `resize`       | `<width>x<height>`

This command will resize an image to the size specified in the image geometry.

The resize image geometry format is `<width>x<height>` which will resize an image while
preserving its aspect ratio. Since the aspect ratio is preserved the resulting image may
be smaller than requested.

If you need an exact size you can add the `!` symbol to your request, such as `100x100!`. This
will not preserve aspect ratios and may cause your image to compress or stretch. Listing 1 shows
how the image can be stretched, Listing 2 shows it without the `!`.

![Listing 1 - resize/100x100!](../assets/resize100x100exclamation.jpg "Listing 1")
<span class="caption">Listing 1 - `/v4/.../resize/100x100!/?url=...` results in a 100x100 image</span>

![Listing 2 - resize/100x100](../assets/resize100x100.jpg "Listing 2")
<span class="caption">Listing 2 - `/v4/.../resize/100x100/?url=...` results in a 80x100 image</span>

> Some geometry formats use symbols that need to be escaped so make sure to always escape
> command arguments.
>
> An example would be resizing by percentage. We're can't have `%` in the url so it needs 
> to be escaped. For example, in `resize/15%25`, the `%25` is the url encoding for `%`.
>