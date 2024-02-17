# Signing Requests

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