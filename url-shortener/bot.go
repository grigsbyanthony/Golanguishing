

package main

import (
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/bwmarrin/discordgo"
)

// prefix for bot commands
const prefix = "!shorten "

func main() {
    // Read bot token from environment variable
    token := os.Getenv("DISCORD_BOT_TOKEN")
    if token == "" {
        log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
    }

    // Create a new Discord session
    dg, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    // Register the messageCreate func as a callback for MessageCreate events.
    dg.AddHandler(messageCreate)

    // Open a websocket connection to Discord
    if err = dg.Open(); err != nil {
        log.Fatalf("Error opening connection to Discord: %v", err)
    }
    defer dg.Close()

    fmt.Println("Discord bot is now running. Press CTRL-C to exit.")
    // Block forever
    select {}
}

// messageCreate is called every time a new message is created on any channel that the bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore messages from the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }

    // Check if the message starts with the command prefix
    if strings.HasPrefix(m.Content, prefix) {
        longURL := strings.TrimSpace(strings.TrimPrefix(m.Content, prefix))
        if longURL == "" {
            s.ChannelMessageSend(m.ChannelID, "Please provide a URL to shorten. Usage: `!shorten <URL>`")
            return
        }

        // Call the existing shorten function from main.go
        shortURL, err := shorten(longURL)
        if err != nil {
            log.Printf("Error shortening URL: %v", err)
            s.ChannelMessageSend(m.ChannelID, "❌ Failed to shorten URL.")
            return
        }

        // Send back the shortened URL
        msg := fmt.Sprintf("🔗 Short URL: %s", shortURL)
        s.ChannelMessageSend(m.ChannelID, msg)
    }
}