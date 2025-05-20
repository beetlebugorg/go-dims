# Configuration

go-dims is configured entirely through environment variables. This makes it portable,
deployment-friendly, and easy to integrate into Docker, AWS Lambda, or other runtime environments.

There are very few required settings — most are optional and designed to give you control over how
images are processed, cached, and served.

## 🚀 Quick Start

You can set configuration values inline when starting the server:

```
DIMS_SIGNING_KEY="mysecret" DIMS_ERROR_BACKGROUND="#ffffff" ./dims serve
```

Or define them in an `.env` file or through your platform’s configuration system (e.g., Docker
Compose, AWS Lambda environment variables, etc.).

## ✅ Required Setting

| Variable           | Description                                              |
|--------------------|----------------------------------------------------------|
| DIMS_SIGNING_KEY   | Secret key used to verify signed URLs. **Required.**     |

## 🔧 Configuration Areas

    Configuration is organized into logical sections. Each section has its own detailed documentation:

- [📌 General Configuration](./general): Logging, error images, timeouts, and runtime behavior.
- [🧠 Cache Control](./cache-control): HTTP cache headers like `Cache-Control`, `Expires`, and `Last-Modified`.
- [🧪 Image Compression](./image-compression): JPEG, PNG, and WebP output tuning.
- [📡 Image Sources](./image-sources): Configure sources like HTTP, S3, and local files.
- [🚚 Migrating from mod_dims](./mod-dims): Migration guide for mod_dims users.

## 💡 Tips

All values are read once during startup.

Booleans should be set as `true` or `false`.

Lists (e.g., excluded formats) are comma-separated: `DIMS_EXCLUDED_OUTPUT_FORMATS=tiff,gif`

If you're deploying go-dims in production, we recommend using environment-specific secrets managers
to safely handle values like `DIMS_SIGNING_KEY`.
