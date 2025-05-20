---
sidebar_position: 4
---

# Migrating from mod_dims

Migrating from `mod-dims` to `go-dims` is straightforward. Most settings have either been renamed for clarity or removed if no longer necessary. This guide outlines the key differences and how to map your existing `dims.conf` settings to environment variables used by `go-dims`.

---

### `DimsClient`

The `DimsClient` directive has been removed. Each of its options is now configured globally via environment variables.

Original `DimsClient` format:

```
<appId> <noImageUrl> <cache control max-age> <edge control downstream-ttl> <trustSource?> <minSourceCache> <maxSourceCache> <password>
```

**Mapping to `go-dims`:**

| mod-dims Field             | go-dims Equivalent                       |
|----------------------------|------------------------------------------|
| `<appId>`                  | No longer used                          |
| `<noImageUrl>`             | `DIMS_ERROR_BACKGROUND`                 |
| `<cache control max-age>`  | `DIMS_CACHE_CONTROL_DEFAULT`            |
| `<edge control downstream>`| `DIMS_EDGE_CONTROL_DOWNSTREAM_TTL`      |
| `<trustSource?>`           | `DIMS_CACHE_CONTROL_USE_ORIGIN`         |
| `<minSourceCache>`         | `DIMS_CACHE_CONTROL_MIN`                |
| `<maxSourceCache>`         | `DIMS_CACHE_CONTROL_MAX`                |
| `<password>`               | `DIMS_SIGNING_KEY`                      |

---

## ✏️ Renamed

| mod-dims                  | go-dims                       |
|---------------------------|-------------------------------|
| `DimsCacheExpire`         | `DIMS_CACHE_CONTROL_DEFAULT`  |
| `DimsNoImageCacheExpire`  | `DIMS_CACHE_CONTROL_ERROR`    |

---

## 📦 Moved

All `DimsImagemagick*` settings have been removed. go-dims does not use Imagemagick.

Removed settings:
- `DimsImagemagickTimeout`
- `DimsImagemagickMemorySize`
- `DimsImagemagickAreaSize`
- `DimsImagemagickMapSize`
- `DimsImagemagickDiskSize`

---

## ❌ Removed

These options are no longer supported in `go-dims`:

- `DimsClient`
- `DimsDefaultImageURL` - Replaced by `DIMS_ERROR_BACKGROUND`
- `DimsAddWhitelist` - This was only used /dims3/ endpoint in mod_dims.
- `DimsDisableEncodedFetch`
- `DimsUserAgentEnabled` - This is now automatic.
- `DimsUserAgentOverride`
- `DimsOptimizeResize` - This is the default behavior in `go-dims`.

---

## 🧪 Example: `dims.conf` to `go-dims`

**Original:**
```
DimsDownloadTimeout 60000
DimsImagemagickTimeout 20000

# DimsClient <appId> <noImageUrl> <cache control max-age> <edge control downstream-ttl> <trustSource?> <minSourceCache> <maxSourceCache> <password>
DimsAddClient TEST http://placehold.it/350x150 604800 604800 trust 604800 604800 t3st

DimsDefaultImageURL http://placehold.it/350x150
DimsCacheExpire 604800
DimsNoImageCacheExpire 60
DimsDefaultImagePrefix

DimsAddWhitelist www.google.com
```

**Migration:**
```
DIMS_SIGNING_KEY=t3st
DIMS_ERROR_BACKGROUND="#cccccc"
DIMS_CACHE_CONTROL_DEFAULT=604800
DIMS_CACHE_CONTROL_ERROR=60
DIMS_EDGE_CONTROL_DOWNSTREAM_TTL=604800
DIMS_CACHE_CONTROL_USE_ORIGIN=true
DIMS_CACHE_CONTROL_MIN=604800
DIMS_CACHE_CONTROL_MAX=604800
DIMS_DOWNLOAD_TIMEOUT=60000
```

You can skip any removed or deprecated options. Most behaviors are now automatic or unnecessary in `go-dims`.

---