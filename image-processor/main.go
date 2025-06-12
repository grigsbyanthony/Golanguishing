// main.go
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"gopkg.in/gographics/imagick.v3/imagick"
)

var uploadFormTmpl = template.Must(template.New("upload").Parse(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Image Filter Tool</title>
</head>
<body>
  <h1>Upload an Image</h1>
  <form enctype="multipart/form-data" action="/upload" method="post">
    <input type="file" name="image" accept="image/*" required><br><br>
    <label><input type="radio" name="filter" value="grayscale" checked> Grayscale</label><br>
    <label><input type="radio" name="filter" value="blur"> Gaussian Blur</label><br><br>
    <!-- Only used if blur is chosen -->
    <label>Radius: <input type="number" name="radius" value="5" min="1"></label>
    <label>Sigma: <input type="number" name="sigma" value="2" min="0.1" step="0.1"></label><br><br>
    <button type="submit">Upload & Process</button>
  </form>
</body>
</html>
`))

func main() {
	// Initialize the ImageMagick environment
	imagick.Initialize()
	defer imagick.Terminate()

	http.HandleFunc("/", serveForm)
	http.HandleFunc("/upload", handleUpload)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// serveForm renders the upload HTML form.
func serveForm(w http.ResponseWriter, r *http.Request) {
	if err := uploadFormTmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// handleUpload receives the uploaded image, applies the selected filter,
// and streams back the result as a downloadable PNG.
func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Image too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file into buffer
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		http.Error(w, "Failed to buffer image", http.StatusInternalServerError)
		return
	}

	// Prepare MagickWand
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image from memory
	if err := mw.ReadImageBlob(buf.Bytes()); err != nil {
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	// Choose filter
	filter := r.FormValue("filter")
	switch filter {
	case "grayscale":
		if err := mw.SetImageType(imagick.IMAGE_TYPE_GRAYSCALE); err != nil {
			http.Error(w, "Failed to convert to grayscale", http.StatusInternalServerError)
			return
		}
	case "blur":
		radius, _ := strconv.Atoi(r.FormValue("radius"))
		sigma, _ := strconv.ParseFloat(r.FormValue("sigma"), 64)
		if radius < 1 {
			radius = 1
		}
		if sigma <= 0 {
			sigma = 1
		}
		if err := mw.GaussianBlurImage(float64(radius), sigma); err != nil {
			http.Error(w, "Failed to apply blur", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unknown filter", http.StatusBadRequest)
		return
	}

	// Set output format to PNG
	if err := mw.SetImageFormat("png"); err != nil {
		http.Error(w, "Failed to set output format", http.StatusInternalServerError)
		return
	}

	// Stream the result back
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="processed.png"`)
	w.Write(mw.GetImageBlob())
}