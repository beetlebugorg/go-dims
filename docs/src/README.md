# Dynamic Image Manipulation Service

Dims is an HTTP microservice for **dynamic image manipulation** written in Go. Use dims
to resize images on-the-fly avoiding the costly process of pre-computing and storing
images.

If you're building a website you probably need a service like dims to resize images
for publishing. With dims you can do that programmatically, on-the-fly.

It can reduce costs by:
- only computing images when you need them.
- not storing image size renditions.

It can improve your development and publishing experience by:
- allowing your developers to easily define image sizes in their frontend code
- allowing your users to publish images without manipulating them first

## License

DIMS User Guide © 2024 by Jeremy Collins is licensed under [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/). 