package main

import (
	"fmt"

	"github.com/lonrun/multi-platform-live-streaming-system/backend/server"
)

func main() {
	s, err := server.NewServer()
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		return
	}

	fmt.Println("Starting server at :8000")
	s.Run(":8000")
}
