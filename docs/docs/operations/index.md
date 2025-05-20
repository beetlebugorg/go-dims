
# Supported Operations

go-dims supports a variety of image transformation operations that can be chained together in the URL path. Below is a quick index — click any operation to learn more.

## 📐 Transformations

| Operation                                     | Description                          |
|-----------------------------------------------|--------------------------------------|
| [`resize`](./transformations/resize.md)       | Resize image dimensions              |
| [`crop`](./transformations/crop.md)           | Crop to specific region              |
| [`thumbnail`](./transformations/thumbnail.md) | Fit within bounding box (smart crop) |
| [`flipflop`](./transformations/flipflop.md)   | Flip horizontally and vertically     |
| [`rotate`](./transformations/rotate.md)       | Rotate by degrees                    |

## 🎨 Adjustments

| Operation                                   | Description              |
|---------------------------------------------|--------------------------|
| [`brightness`](./adjustments/brightness.md) | Adjust brightness        |
| [`autolevel`](./adjustments/autolevel.md)   | Auto-level contrast      |
| [`sharpen`](./adjustments/sharpen.md)       | Apply sharpening filter  |
| [`grayscale`](./adjustments/grayscale.md)   | Convert to grayscale     |
| [`invert`](./adjustments/invert.md)         | Invert colors            |

## 🧾 Output Options

| Operation                        | Description                        |
|----------------------------------|------------------------------------|
| [`format`](./output/format.md)   | Set output format (jpg, png, etc.) |
| [`quality`](./output/quality.md) | Adjust compression quality         |
| [`strip`](./output/strip.md)     | Strip metadata                     |

## 💧 Special

| Operation                             | Description         |
|---------------------------------------|---------------------|
| [`watermark`](./special/watermark.md) | Overlay a watermark |

---

> Want to learn more? Click any operation above to view usage examples, parameter options, and gotchas.