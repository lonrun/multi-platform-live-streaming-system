package main

import (
	"fmt"
	"net/http"

	"github.com/lonrun/multi-platform-live-streaming-system/backend/server"
)

func main() {
	go func() {
		s, err := server.NewServer()
		if err != nil {
			fmt.Printf("Error creating server: %v\n", err)
			return
		}

		fmt.Println("Starting server at :8000")
		s.Run(":8000")
	}()

	go func() {
		fs := http.FileServer(http.Dir("../frontend"))
		http.Handle("/", fs)
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Printf("Error starting frontend server: %v\n", err)
			return
		}
	}()

	select {}
}
