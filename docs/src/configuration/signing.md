# Request Signing

All image manipulation requests are required to be signed with a secret
shared between your web framework, and dims.

If you encounter key mismatch errors ensure that both the signing key and signing algorithm are the
same between the client, and dims.

### DIMS_SIGNING_KEY

*The `DIMS_SIGNING_KEY` is used to verify the signature of each image manipulation request.*

Make sure to set this to a secure random value with at least 32 characters. Use your favorite
password manager to generate something strong.

> Keep this value a secret. Don't check it into revision control. You'll
> need it to sign your dims image urls.

**This setting is required.**

### DIMS_SIGNING_ALGORITHM

*This sets the algorithm used for signing image manipulation requests.*

The default is `hmac-sha256`.

You generally shouldn't need to touch this setting. However, if you previously used
mod-dims you can set this to `md5` for backward compatibility.

Changing this will cause all signatures to change. This will most likely bust whatever
cache you have in from of dims. Be careful.

This setting also affects places where dims needs as hash. For example, `Etag` headers
will be generated with `sha256` by default. Any places that use hash algorithms will
use the same algorithm as signing.

> The md5 algorithm is old and broken. It's highly recommended to use the new
> `hmac-sha256` algorithm.