

# DOCUMENTATION.md

This document provides a line-by-line explanation of the code in `main.go` and `bot.go`.

---

## File: main.go

```go
package main                              // Defines this file as part of the main package; the entry point for the Go program.

import (                                  // Import block for required standard libraries.
    "encoding/json"                      // For JSON encoding/decoding of URL mappings.
    "flag"                               // For parsing CLI flags.
    "fmt"                                // For formatted I/O.
    "log"                                // For logging errors and informational messages.
    "math/rand"                          // For generating random codes.
    "net/http"                           // For HTTP server functionality.
    "os"                                 // For file and environment interactions.
    "sync"                               // For concurrency-safe locks.
    "time"                               // For seeding randomness and timestamps.
)

const (                                  // Constants used throughout the program.
    dbFile     = "urls.json"             // Filename for persisting URL mappings.
    baseURL    = "http://localhost:8080/"// Base URL for generating short links.
    codeLength = 6                       // Length of randomly generated code.
)

var (                                   // Global variables.
    mu   sync.RWMutex                   // Read/Write mutex to protect the `urls` map.
    urls map[string]string              // In-memory map from code => original URL.
)

func init() {                           // init runs before main.
    rand.Seed(time.Now().UnixNano())   // Seed RNG with current time for uniqueness.
    urls = make(map[string]string)     // Initialize the map.
    if err := load(); err != nil {     // Attempt to load persisted mappings.
        log.Println("Failed to load DB:", err) // Log but do not exit on load error.
    }
}

// load reads URL mappings from disk if the file exists.
func load() error {
    file, err := os.Open(dbFile)             // Open the JSON file.
    if os.IsNotExist(err) {                  // If file not found, nothing to load.
        return nil
    } else if err != nil {                   // If other error, propagate.
        return err
    }
    defer file.Close()                       // Ensure file is closed.
    return json.NewDecoder(file).Decode(&urls) // Decode JSON into the map.
}

// save writes current URL mappings to disk safely.
func save() error {
    temp := dbFile + ".tmp"                   // Write to temp file first.
    file, err := os.Create(temp)              // Create temp file.
    if err != nil {
        return err
    }
    encoder := json.NewEncoder(file)          // JSON encoder.
    encoder.SetIndent("", "  ")               // Pretty-print with indentation.
    if err := encoder.Encode(urls); err != nil { // Write map to temp file.
        file.Close()
        return err
    }
    file.Close()                              // Close temp file.
    return os.Rename(temp, dbFile)            // Atomically replace the old file.
}

// generateCode produces a random alphanumeric string of length codeLength.
func generateCode() string {
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    b := make([]rune, codeLength)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))] // Pick a random rune each iteration.
    }
    return string(b)
}

// shorten generates a unique code for a URL and persists it.
func shorten(u string) (string, error) {
    mu.Lock()                  // Acquire write lock.
    defer mu.Unlock()          // Release lock when done.
    for {
        code := generateCode() 
        if _, exists := urls[code]; !exists { // Ensure code uniqueness.
            urls[code] = u        // Store mapping in memory.
            if err := save(); err != nil { // Persist to disk.
                return "", err
            }
            return baseURL + code, nil // Return full short URL.
        }
    }
}

// redirectHandler handles GET /{code} and redirects if found.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Path[1:]      // Trim leading slash to get code.
    mu.RLock()                  // Acquire read lock.
    defer mu.RUnlock()          // Release after reading.
    if dest, ok := urls[code]; ok {      // Lookup original URL.
        http.Redirect(w, r, dest, http.StatusFound) // Redirect client.
    } else {
        http.NotFound(w, r)     // Return 404 if code not found.
    }
}

// shortenHandler handles POST /shorten with JSON body {"url":"..."}.
func shortenHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {           // Enforce POST only.
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req struct {
        URL string ` + "`json:\"url\"`" + ` // Request payload struct.
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // Decode JSON.
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    short, err := shorten(req.URL)             // Shorten the URL.
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    resp := map[string]string{"short_url": short} // Prepare response.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)             // Send JSON response.
}

// runServer wires up HTTP handlers and starts listening.
func runServer() {
    http.HandleFunc("/", redirectHandler)
    http.HandleFunc("/shorten", shortenHandler)
    log.Println("Starting server at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// runCLI handles command-line flags for either serving or shortening.
func runCLI(args []string) {
    fs := flag.NewFlagSet("urls", flag.ExitOnError)
    longURL := fs.String("url", "", "URL to shorten") // -url flag.
    serve := fs.Bool("serve", false, "Run HTTP server") // -serve flag.
    fs.Parse(args)

    if *serve {
        runServer()             // Launch HTTP server if requested.
        return
    }
    if *longURL != "" {        // If URL provided, shorten via CLI.
        short, err := shorten(*longURL)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println("Shortened URL:", short)
        return
    }
    fs.Usage()                  // Show usage if no flags.
}

func main() {
    if len(os.Args) > 1 {       // If any CLI args present
        runCLI(os.Args[1:])     // handle CLI mode
    } else {
        runServer()             // else default to server mode
    }
}
```

---

## File: bot.go

```go
package main                            // Same package so it can reuse shorten().

import (
    "fmt"                              // For string formatting.
    "log"                              // For logging errors.
    "os"                               // To read environment variables.
    "strings"                          // For string manipulation.
    "github.com/bwmarrin/discordgo"    // DiscordGo library.
)

// prefix for identifying bot commands
const prefix = "!shorten "

func main() {
    token := os.Getenv("DISCORD_BOT_TOKEN")  // Read bot token.
    if token == "" {
        log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
    }

    dg, err := discordgo.New("Bot " + token) // Create Discord session.
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    dg.AddHandler(messageCreate)            // Register message handler.

    if err = dg.Open(); err != nil {        // Open WebSocket to Discord.
        log.Fatalf("Error opening connection to Discord: %v", err)
    }
    defer dg.Close()                        // Ensure session closes on exit.

    fmt.Println("Discord bot is now running. Press CTRL-C to exit.")
    select {}                               // Block forever.
}

// messageCreate handles incoming Discord messages.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {     // Ignore the bot's own messages.
        return
    }

    if strings.HasPrefix(m.Content, prefix) { // Check command prefix.
        longURL := strings.TrimSpace(strings.TrimPrefix(m.Content, prefix))
        if longURL == "" {
            s.ChannelMessageSend(m.ChannelID,
                "Please provide a URL to shorten. Usage: `!shorten <URL>`")
            return
        }

        shortURL, err := shorten(longURL)  // Call shorten() from main.go.
        if err != nil {
            log.Printf("Error shortening URL: %v", err)
            s.ChannelMessageSend(m.ChannelID, "❌ Failed to shorten URL.")
            return
        }

        msg := fmt.Sprintf("🔗 Short URL: %s", shortURL)
        s.ChannelMessageSend(m.ChannelID, msg) // Send shortened link.
    }
}
```