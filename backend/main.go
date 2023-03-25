package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/user/multi-platform-live-streaming-system/backend/server"
)

func main() {
	http.HandleFunc("/", server.HandleConnections)

	fmt.Println("Starting server at :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err.Error())
	}
}
