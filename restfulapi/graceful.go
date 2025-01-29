package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type Response struct {
    Message string `json:"message"`
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := Response{Message: "Hello, JSON World!"}
    json.NewEncoder(w).Encode(response)
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/json", jsonHandler)

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // Capture system signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        log.Println("Server running on http://localhost:8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    <-quit
    log.Println("\nShutting down server...")

    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
    defer shutdownCancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited cleanly")
}
