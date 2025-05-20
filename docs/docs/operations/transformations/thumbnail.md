---
sidebar_position: 3
---

# Thumbnail

Generate a thumbnail version of the image, fitting it into the specified bounding box.

## Syntax

| Command     | Argument Format    |
|-------------|--------------------|
| `thumbnail` | `<width>x<height>` |

## Behavior

- Resizes the image to **fill** the specified box, cropping as needed to preserve aspect ratio.
- Automatically strips metadata from the output.
- Ideal for generating consistent-sized thumbnails from varying source images.

## Example

Compare a standard resize with a thumbnail resize on the same image:

#### Thumbnail to 200×200 — image is cropped to fill exact box.

```
/v5/thumbnail/200x200/?url=pexels-photo-1539116.jpeg
```

![Thumbnail Example](../../assets/thumbnail200x200.jpg)


#### Resize to 200×200 — image fits within bounds, but may not fill box completely.

```
/v5/resize/200x200/?url=pexels-photo-1539116.jpeg
```

![Resize Example](../../assets/resize200x200.jpg)

