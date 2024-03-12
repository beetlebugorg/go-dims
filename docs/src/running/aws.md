# Amazon Web Services

Serverless operation is supported for AWS Lambda Functional URL, and S3 Object Lambda.

Each mode offers different trade-offs. 

The Lambda Functional URL mode allows you to resize images from any HTTP source, but also exposes
that source to the client. This may not be ideal such as in Digital Asset Management systems.

The S3 Object Lambda mode doesn't expose the origin and allows you to keep your S3 bucket completely
private. However, it only works with S3 as the source.

## AWS S3 Object Lambda



```
CloudFront -> S3 Object Lambda Access Point -> S3 Access Point -> S3
```

## AWS Lambda Functional URL

```
CloudFront -> Lambda Functional URL -> Image Source (http, s3, etc)
```
