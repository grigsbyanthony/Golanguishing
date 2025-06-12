

package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "sync"
    "time"
)

const (
    dbFile     = "urls.json"
    baseURL    = "http://localhost:8080/"
    codeLength = 6
)

var (
    mu   sync.RWMutex
    urls map[string]string
)

func init() {
    rand.Seed(time.Now().UnixNano())
    urls = make(map[string]string)
    if err := load(); err != nil {
        log.Println("Failed to load DB:", err)
    }
}

// load reads the URL mappings from the JSON file.
func load() error {
    file, err := os.Open(dbFile)
    if os.IsNotExist(err) {
        return nil
    } else if err != nil {
        return err
    }
    defer file.Close()
    return json.NewDecoder(file).Decode(&urls)
}

// save writes the URL mappings to the JSON file.
func save() error {
    temp := dbFile + ".tmp"
    file, err := os.Create(temp)
    if err != nil {
        return err
    }
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(urls); err != nil {
        file.Close()
        return err
    }
    file.Close()
    return os.Rename(temp, dbFile)
}

// generateCode produces a random string of length codeLength.
func generateCode() string {
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    b := make([]rune, codeLength)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

// shorten creates a new short code for the given URL and persists it.
func shorten(u string) (string, error) {
    mu.Lock()
    defer mu.Unlock()
    for {
        code := generateCode()
        if _, exists := urls[code]; !exists {
            urls[code] = u
            if err := save(); err != nil {
                return "", err
            }
            return baseURL + code, nil
        }
    }
}

// redirectHandler looks up the code and redirects if found.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Path[1:]
    mu.RLock()
    defer mu.RUnlock()
    if dest, ok := urls[code]; ok {
        http.Redirect(w, r, dest, http.StatusFound)
    } else {
        http.NotFound(w, r)
    }
}

// shortenHandler accepts a JSON body with {"url": "..."} and responds with {"short_url": "..."}.
func shortenHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req struct {
        URL string `json:"url"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    short, err := shorten(req.URL)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    resp := map[string]string{"short_url": short}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// runServer sets up the HTTP handlers and starts listening.
func runServer() {
    http.HandleFunc("/", redirectHandler)
    http.HandleFunc("/shorten", shortenHandler)
    log.Println("Starting server at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// runCLI parses flags for either serving or shortening via command-line.
func runCLI(args []string) {
    fs := flag.NewFlagSet("urls", flag.ExitOnError)
    longURL := fs.String("url", "", "URL to shorten")
    serve := fs.Bool("serve", false, "Run HTTP server")
    fs.Parse(args)

    if *serve {
        runServer()
        return
    }
    if *longURL != "" {
        short, err := shorten(*longURL)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println("Shortened URL:", short)
        return
    }
    fs.Usage()
}

func main() {
    if len(os.Args) > 1 {
        runCLI(os.Args[1:])
    } else {
        runServer()
    }
}