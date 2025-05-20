---
sidebar_position: 2
---

# Cache Control

Proper cache control is essential for performance, cost efficiency, and SEO. It directly impacts
your server load, CDN effectiveness, and Lighthouse scores ([why it
matters](https://developer.chrome.com/docs/lighthouse/performance/uses-optimized-images#how_lighthouse_flags_images_as_optimizable)).

The default settings in `go-dims` are based on over a decade of production use and are optimized for
high-traffic sites with dynamic image transformations. If your use case is more static or tightly
controlled, you may want to adjust them.

---

## `DIMS_CACHE_CONTROL_DEFAULT`

Sets the default `max-age` (in seconds) for the `Cache-Control` header on successful image responses.

- **Default:** `31536000` (1 year)

A long cache lifetime is safe because every image variant is URL-based and includes a signature — changes to transformations automatically bust caches.

> Previously known as `DimsCacheExpire` in mod-dims.

---

## `DIMS_CACHE_CONTROL_ERROR`

Sets the `max-age` (in seconds) for responses that return an error image.

- **Default:** `60`

This helps prevent excessive retries while still allowing for timely recovery when transient issues are resolved.

Example:

```
$ DIMS_CACHE_CONTROL_ERROR=11 DIMS_DEVELOPMENT_MODE=true ./dims serve
```

Then simulate a bad request:

```
curl -v "http://127.0.0.1:8080/dims4/default/1/resize/1x1/?url=1"
```

Response:
```
< HTTP/1.1 400 Bad Request  
< Cache-Control: max-age=11, public  
< Expires: [11 seconds after request]  
```

> Previously known as `DimsNoImageCacheExpire` in mod-dims.

---

## `DIMS_CACHE_CONTROL_USE_ORIGIN`

Delegates cache duration to the `Cache-Control` header set by the origin server.

- **Default:** `false`

Useful if you control the origin (e.g., S3) and want fine-grained control on a per-image basis.

> ⚠️ Only enable this if you fully control the origin.  
> If the origin sets an inappropriately long or short `max-age`, it could lead to excessive load or stale content.

---

## `DIMS_CACHE_CONTROL_MIN`

Defines the **minimum** `max-age` when using origin-based caching.

- **Default:** `0` (no minimum)

This acts as a safety net — if the origin sets too low a value, this will override it.

Only takes effect when `DIMS_CACHE_CONTROL_USE_ORIGIN=true`.

---

## `DIMS_CACHE_CONTROL_MAX`

Defines the **maximum** `max-age` when using origin-based caching.

- **Default:** `0` (no maximum)

This caps the TTL that can be inherited from the origin. Useful to prevent overly aggressive caching.

Only takes effect when `DIMS_CACHE_CONTROL_USE_ORIGIN=true`.

---