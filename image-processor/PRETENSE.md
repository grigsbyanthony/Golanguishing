# PRETENSE.md

## What I Learned, I Think

- **HTTP Servers in Go**: How to set up routes, handle requests, and serve HTML forms using the `net/http` package.
- **Multipart Form Parsing**: Reading uploaded files directly from an HTTP request without saving to disk.
- **Go Templates**: Rendering dynamic HTML using `html/template`.
- **ImageMagick Integration**: Initializing the ImageMagick environment and manipulating images through the Go binding (`imagick`).
- **In-Memory Processing**: Loading images into memory, applying filters, and streaming results back to the client.
- **HTTP Response Headers**: Prompting file downloads by setting `Content-Type` and `Content-Disposition` headers.

## Use Cases

- **Simple Image Filter Service**: Quickly add grayscale or blur functionality to any web application.
- **Microservices**: Deploy as a Dockerized service for on-demand image processing in larger systems.
- **Prototyping**: Test new image transformations before integrating into production pipelines.
- **Educational Tool**: Demonstrate core Go and ImageMagick concepts to newcomers.
- **Bot Extensions**: Integrate into chat platforms (Discord, Slack) to process and return images in real time.
