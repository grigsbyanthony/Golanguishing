

# DOCUMENTATION.md

This tool is a Go HTTP server that allows you to upload an image, apply filters (grayscale or Gaussian blur), and download the processed image without writing to disk.

## Prerequisites

- Go 1.18 or newer
- [ImageMagick](https://imagemagick.org) installed on your system
- Go binding for ImageMagick:
  ```bash
  go get gopkg.in/gographics/imagick.v3/imagick
  ```

## Installation

1. Install ImageMagick on macOS with Homebrew:
   ```bash
   brew install imagemagick
   ```
2. Fetch the Go binding:
   ```bash
   go get gopkg.in/gographics/imagick.v3/imagick
   ```

## Running the Server

```bash
go run main.go
```

Then open your browser and navigate to `http://localhost:8080`.

## Endpoints

### `GET /`
Serves an HTML form for uploading an image and selecting a filter.

### `POST /upload`
- Parses the uploaded multipart form containing the image and filter parameters.
- Reads the image into memory and loads it into a `MagickWand`.
- Applies the chosen filter:
  - **Grayscale**: Converts the image to grayscale.
  - **Gaussian Blur**: Applies a blur using `GaussianBlurImage(radius, sigma)`.
- Sets the output format to PNG and streams the processed image back with a download prompt.

## Code Overview

- `main()`:
  - Calls `imagick.Initialize()` and `imagick.Terminate()` to manage the ImageMagick environment.
  - Registers handlers for `/` (HTML form) and `/upload` (processing logic).  
- `serveForm(w, r)`:
  - Renders the HTML upload form using a `template.Template`.
- `handleUpload(w, r)`:
  1. Parses the multipart form and reads the uploaded file into a buffer.
  2. Loads the image into a `MagickWand` from the buffered bytes.
  3. Applies the selected filter based on `r.FormValue("filter")`.
  4. Sets the image format to PNG.
  5. Writes the image blob to the HTTP response with appropriate headers.