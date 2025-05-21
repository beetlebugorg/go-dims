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

---

## `DIMS_SIGNING_KEY`

**This setting is required.**

This key is used to validate every incoming image request. If the signature doesn’t match, the
request will be rejected.

:::warning

Never expose or commit this value to source control.  
Treat it like a production secret — store it in a secure environment variable, secret manager, or encrypted config.

:::

:::tip Best Practice

Use at least 32 characters of high-entropy random data - Generate using your password manager or a secure CLI tool

:::