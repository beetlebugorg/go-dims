# Configuration

You can configure dims through environment variables. 

```bash
$ DIMS_SIGNING_KEY="mysecret" DIMS_ERROR_BACKGROUND="#ffffff" ./dims serve
```

There are not many settings, and only one required setting (`DIMS_SIGNING_KEY`). Most
of the settings configure how dims returns cache control headers.
