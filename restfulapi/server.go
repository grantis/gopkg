package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Response struct {
    Message string `json:"message"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hi Mum!")
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := Response{Message: "Hi Mum, its JSON!"}
    json.NewEncoder(w).Encode(response)
}

func main() {
    http.HandleFunc("/", helloHandler)
    http.HandleFunc("/json", jsonHandler)

    fmt.Println("Server running on http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}
