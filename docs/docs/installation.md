---
sidebar_position: 2
---

# Installation 

Welcome to go-dims — a high-performance, production-grade image processing service powered by
libvips. It’s fast, lightweight, and designed to drop seamlessly into your stack. You can run
go-dims in one of three simple ways: 
- 🐳 as a Docker container
- ☁️ as an AWS Lambda function
- 📦 as a standalone binary

This guide will walk you through each option so you can get up and running quickly — whether you’re
developing locally or deploying to the cloud. Let’s dive in. 👇

## 🐳 Docker

The fastest way to get started with go-dims is by running it in a Docker container. This gives you a
clean, self-contained environment, no setup required.

To start go-dims in development mode on port 8080, run:

```shell
$ docker run \
    -e DIMS_SIGNING_KEY=devmode \
    -e DIMS_DEVELOPMENT_MODE=true \
    -p 8080:8080 ghcr.io/beetlebugorg/go-dims serve
```

🛠️ Development mode disables signature verification, so the signing key can be anything — but it still must be provided.

Once the container is running, you can confirm the service is alive:

```shell
❯ curl http://127.0.0.1:8080/dims-status
ALIVE
```

Then, open this URL in your browser to test image processing:

```shell
http://127.0.0.1:8080/v5/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg
```

### Image tags

| Tag | Contents |
|---|---|
| `latest` | The most recent tagged release. Only a release moves it. |
| `<version>` | A specific release, for example `0.4.1`. Never moves. |
| `next` | The most recent release candidate, a tag such as `v1.0.0-rc1`. |
| `edge` | The most recent rebuild from `main`, published by hand. |
| `<short sha>` | The exact commit a rebuild was made from. Never moves. |

Pin a version in production. `latest`, `next`, and `edge` all move, so each one changes under you.

## 📦 Standalone Binary

Install the `dims` command with Homebrew, on macOS or on Linux:

```shell
$ brew install beetlebugorg/tap/go-dims
```

Confirm the install:

```shell
❯ dims version
1.0.0
```

The command runs the server and signs URLs:

```shell
$ dims serve
$ dims sign --key-file=signing-key.txt "http://localhost:8080/v5/resize/200x200/?url=cat.jpg"
```

The formula tracks the newest tag, a release candidate such as `v1.0.0-rc1` included. Run `dims version` to see which one you have.

Every release also attaches the binary directly. Download `dims-<platform>-<arch>.zip` from the [releases page](https://github.com/beetlebugorg/go-dims/releases), unzip it, and run `dims`. The Linux build is static and needs nothing at run time. The macOS build links libvips, which Homebrew installs for you.

## ☁️ AWS Lambda

go-dims can be deployed as a compact, production-ready AWS Lambda function — ideal for real-time
image processing without the burden of managing servers, scaling infrastructure, or handling idle
capacity. With Lambda’s “pay only for what you use” model, you get automatic scaling, built-in fault
tolerance, and zero cost when idle — making it a perfect fit for dynamic, bursty workloads like
on-the-fly image transformation.

This setup uses Lambda Function URLs to expose your function over HTTPS, making it easy to integrate
with websites, apps, or CDNs without any additional API Gateway setup.

### Prerequisites

Before you deploy go-dims to AWS Lambda, make sure you have the following in place:
- An AWS account
- The AWS CLI installed and configured with credentials (aws configure)

You’ll also need basic permission to create Lambda functions and function URLs. Make sure your
AWS identity has access to:
- `lambda:CreateFunction`
- `lambda:CreateFunctionUrlConfig`
- `iam:PassRole`
- `logs:* (for basic logging support)`

Once your environment is ready, continue with the deployment steps below.

### Deployment Steps

**Step 1: Download the Lambda Bundle**

Go to the Releases page and download the appropriate prebuilt ZIP file:
- lambda-linux-arm64.zip (recommended for Graviton-based functions)
- lambda-linux-amd64.zip (for x86_64 environments)

**Step 2: Create the IAM Role**

Before you can create the Lambda function, you’ll need an IAM role that allows the function to write logs to CloudWatch.

Create the role and attach the basic execution policy:

```bash
aws iam create-role \
    --role-name go-dims-lambda \
    --assume-role-policy-document '{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": {
            "Service": "lambda.amazonaws.com"
          },
          "Action": "sts:AssumeRole"
        }
      ]
    }'
```

Then, attach the `AWSLambdaBasicExecutionRole` policy to the role. This policy allows the Lambda
function to write logs to CloudWatch.

```bash
aws iam attach-role-policy \
    --role-name go-dims-lambda \
    --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

Once created, grab the role ARN using this command:

```bash
aws iam get-role --role-name go-dims-lambda --query 'Role.Arn' --output text
```

**Step 3: Create the Lambda Function**

Use the downloaded ZIP and the IAM role ARN to create your function:

```bash
aws lambda create-function \
    --function-name go-dims \
    --runtime provided.al2023 \
    --role arn:aws:iam::123456789012:role/your-execution-role \
    --memory-size 512 \
    --handler bootstrap \
    --architectures arm64 \
    --zip-file fileb://lambda-arm64.zip \
    --environment Variables="{DIMS_SIGNING_KEY=devmode,DIMS_DEVELOPMENT_MODE=true}"
```

Replace the `--role` value with the ARN of an IAM role from Step 2. For production deployments, you should
increase the memory size to increase the CPU available to the function.

**Step 4: Create the Function URL**

Create a public HTTPS endpoint using Lambda Function URLs:

```bash
aws lambda add-permission \
  --function-name go-dims \
  --statement-id public-access \
  --action lambda:InvokeFunctionUrl \
  --principal "*" \
  --function-url-auth-type NONE
```

```bash
aws lambda create-function-url-config \
    --function-name go-dims \
    --auth-type NONE \
    --invoke-mode RESPONSE_STREAM
```

This will return a FunctionUrl that you can use in a browser or with any HTTP client. 

:::note

⚠️ Your function is publicly accessible. Production deployments should be deployed behind a CDN such
as CloudFront to cache image resize requests.

:::

**Step 5: Test the Function**

To verify everything is working, open your function URL in a browser:

```shell
https://your-function-url/v5/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg
```