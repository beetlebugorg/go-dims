# mod-dims

The mod-dims and go-dims configurations are largely the same. go-dims gives
some settings better names, and removes unnecessary settings.

This page documents those differences and how to migrate. It's mostly just naming.

### New

Since go-dims has a new signing algorithm we added `DIMS_SIGNING_ALGORITHM`.  Set this to `md5` for
compatiblity with mod-dims.

### Changed

`DimsClient` was removed. Each option it had is now available globally.

The `DimsClient` setting had a lot of options:

    <appId> <noImageUrl> <cache control max-age> <edge control downstream-ttl> <trustSource?> <minSourceCache> <maxSourceCache> <password>

Here is how they map to go-dims settings:

- `<appId>` is now `DIMS_CLIENT_ID`
- `<noImageUrl>` is now `DIMS_ERROR_BACKGROUND`
- `<cache control max-age>` is now `DIMS_CACHE_CONTROL_DEFAULT`
- `<edge control downstream-ttl>` is now `DIMS_EDGE_CONTROL_DOWNSTREAM_TTL`
- `<trustedSource>` is now `DIMS_CACHE_CONTROL_USE_ORIGIN`
- `<minSourceCache>` is now `DIMS_CACHE_CONTROL_MIN`
- `<maxSourceCache>` is now `DIMS_CACHE_CONTROL_MAX`
- `<password>` is now `DIMS_SIGNING_KEY`

### Renamed

- `DimsCacheExpire` to `DIMS_CACHE_CONTROL_DEFAULT`
- `DimsNoImageCacheExpire` to `DIMS_CACHE_CONTROL_ERROR`

### Moved

All settings related to Imagemagick were removed. They didn't disappear though.  You can set
everything mod-dims could using Imagemagick environment variables.

See the Imagemagick documentation for [settings](https://imagemagick.org/script/resources.php#environment).

- `DimsImagemagickTimeout`
- `DimsImagemagickMemorySize`
- `DimsImagemagickAreaSize`
- `DimsImagemagickMapSize`
- `DimsImagemagickDiskSize`

### Removed

- `DimsDefaultImageURL`
- `DimsAddWhitelist`
- `DimsDisableEncodedFetch`
- `DimsUserAgentEnabled`
- `DimsUserAgentOverride`
- `DimsOptimizeResize`

### Example: dims.conf

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
