# PRETENSE.md
This project provided me with some hands-on experience with several of Go’s core strengths and common use cases:

1. **Effortless Web Services**  
   Go’s built-in `net/http` package allows you to spin up a fully functional HTTP server in just a few lines of code. Handling routes, parsing JSON requests, and performing redirects all felt immediate and intuitive.

2. **Concurrent-Safe Data Handling**  
   By using `sync.RWMutex`, I learned how Go makes it simple to protect shared data structures against race conditions. This reinforces Go’s philosophy of easy-to-use, first-class concurrency primitives.

3. **Rich Standard Library**  
   From command-line parsing (`flag`) to JSON serialization (`encoding/json`) and file operations (`os`), Go’s batteries-included approach meant I could build both a CLI tool and a web service without pulling in many third-party packages.

4. **Versatile CLI & Daemon Modes**  
   Combining a `runCLI` function with an HTTP server in the same codebase showed how Go is well-suited for applications that need both interactive command-line tools and long-running network services.

5. **Modular Code Reuse**  
   Splitting the core shortening logic into reusable functions that power both `main.go` and `bot.go` taught me about organizing Go modules and importing code across multiple entry points.

6. **Deployment Simplicity**  
   Building a single, static binary highlighted Go’s ease of deployment for microservices and bots. I can push the compiled executable anywhere—VPS, container, or serverless—and have minimal runtime dependencies.