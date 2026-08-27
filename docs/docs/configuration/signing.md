--- 
sidebar_position: 5 
---

# Signing

All image manipulation requests in `go-dims` must be signed using a shared secret. This ensures URLs
cannot be tampered with or abused, and that your image cache remains effective and secure.

Your web application or image URL generator must use the same signing key and algorithm as `go-dims`
to generate valid URLs.

If you encounter signature mismatch errors, double-check that: 
- The signing key matches exactly on both sides 
- The signing algorithm is consistent

The construction is described in [`/v5`](../endpoints/dims5.md#-signing), with an implementation in four languages.

---

## `DIMS_SIGNING_KEY`

**This setting is required.**

This key is used to validate every incoming image request. If the signature doesn’t match, the
request will be rejected.

This key is also used to decrypt the `eurl` query parameter. For mod_dims compatibility, prepend
`sha1:` to the key.

:::tip

Never expose or commit this value to source control. Treat it like a production secret — store it in
a secure environment variable, secret manager, or encrypted config.

Use at least 32 characters of high-entropy random data - Generate using your password manager or a secure CLI tool

:::

---

## `DIMS_SIGNING_COMPAT`

Accepts the mod_dims signature, where only the parameters named by `_keys` took part.

- **Default:** *(empty)*

Set it to `legacy` while you re-sign existing URLs:

```
DIMS_SIGNING_COMPAT=legacy
```

`go-dims` logs a warning for every request that validates this way, so you can watch the count fall to zero. Clear the setting once it does.

:::warning

The canonical message changes every signature. A URL signed by an earlier `go-dims` release fails, and this setting does not accept it. Re-sign every URL before you upgrade.

:::

:::warning

Legacy mode leaves every parameter outside `_keys` unprotected. That includes `overlay`, so a caller holding one valid signed watermark URL can point the overlay at any address the service can reach. Treat this setting as a migration aid with an end date.

:::
