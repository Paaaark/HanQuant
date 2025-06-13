package main

import (
	"log"
	"net/http"

	"github.com/Paaaark/hanquant/internal/server"
)

func main() {
    srv := server.NewHTTPServer()
    log.Println("Server started at :8081")
    log.Fatal(http.ListenAndServe(":8081", srv))
}
