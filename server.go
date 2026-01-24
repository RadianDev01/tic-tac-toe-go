package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Serve static files from the current directory
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := "8080"
	fmt.Printf("Starting Tic Tac Toe web server on http://localhost:%s\n", port)
	fmt.Println("Open your browser and navigate to the URL above to play!")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
