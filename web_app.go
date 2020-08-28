package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func handler2(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love2 %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/edit", handler2)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
