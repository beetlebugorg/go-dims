# go-dims

Dims is an HTTP microservice for dynamic image manipulation written in Go. Use dims to resize images on-the-fly avoiding the costly process of pre-computing and storing images.

`go-dims` is a Go implementation of the DIMS API as implemented by [mod-dims](https://github.com/beetlebugorg/mod_dims).

If you're building a website you probably need a service like dims to resize images for publishing. With dims you can do that programmatically, on-the-fly.

It can reduce costs by:
- only computing images when you need them.
- not storing image size renditions.

It can improve your development and publishing experience by:
- allowing your developers to easily define image sizes in their frontend code
- allowing your users to publish images without manipulating them first

## Running

Running in development mode disables signature verification:

```
docker run -p 8080:8080 ghcr.io/beetlebugorg/go-dims serve --dev --debug --bind ":8080"
```