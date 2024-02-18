# Cache Control Headers

Cache control can be confusing but it's also super important to get it right. It impacts
not only the load on your dims servers but it also affects your [Lighthouse
scores](https://developer.chrome.com/docs/lighthouse/performance/uses-optimized-images#how_lighthouse_flags_images_as_optimizable).

The defaults used by dims have been used in production for over 10 years. They have been optimized
for good Lighthouse scores. These settings are for high traffic sites with lots of image resizing.
If that's not your use case, these settings may not be right for you.

### DIMS_CACHE_CONTROL_DEFAULT

Use this to set how long clients should cache manipulated images. 

This is the `max-age` in seconds in the `Cache-Control` header.

This default is `31536000` seconds. 

That's **one year**, which is basically *forever*. That's ok though, cache busting
is implicit in the URL so any image manipulation changes the signature, busting the cache!

The was `DimsCacheExpire` in mod-dims.

### DIMS_CACHE_CONTROL_ERROR

Use this to set how long to cache error images.

This is the `max-age` in seconds in the `Cache-Control` header.

This default is `60` seconds. 

You don't want to beat up dims serving generated error images, but you also want
clients to retry and hopefully get a valid image.

Let's see how this manifests. First start up dims, setting the expire to `10` seconds.

```shell
$ DIMS_CACHE_CONTROL_ERROR=11 ./dims serve 
```

Now let's simulate a bad request:

```shell
❯ curl -v "http://127.0.0.1:8080/dims4/default/1/resize/1x1/?url=1"
> GET /dims4/default/1/resize/1x1/?url=1 HTTP/1.1
> 
< HTTP/1.1 500 Internal Server Error
< Cache-Control: max-age=11, public
< Content-Length: 0
< Content-Type: image/jpeg
< Expires: Wed, 14 Feb 2024 00:00:11 GMT
< Date: Wed, 14 Feb 2024 00:00:00 GMT
```

This error image will expire after 11 seconds.

The was `DimsNoImageCacheExpire` in mod-dims.

### DIMS_CACHE_CONTROL_USE_ORIGIN

Use this to delegate `max-age` to the origin.

The default is `false`.

You can use this to exercise more control over cache control headers.  For example if you're storing
images in S3 you can can set cache-control headers for individual files.

> You should only set this to true if you control the origin for your images,
> and you have a need for more fine grained control over max-age.
>
> If you do not control the origin then setting this to true puts you at risk that 
> the origin owner sets max-age too high or two low for your needs. Too low and you
> risk DDOSing yourself.

### DIMS_CACHE_CONTROL_MIN

Use this to set a minimum age from trusted origins.

The default is `0`.

This only has an effect when `DIMS_CACHE_CONTROL_USE_ORIGIN` is set to `true`.

If you trust the origin but you still need a bit of a protection from
misconfiguration or misuse, this is the setting for you.

### DIMS_CACHE_CONTROL_MAX

Use this to set a maximum age from trusted origins.

The default is `0`. This disables maximum value checks.

This only has an effect when `DIMS_CACHE_CONTROL_USE_ORIGIN` is set to `true`.

