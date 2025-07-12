package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve")
	flag.Parse()

	// Get absolute path
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	// Create file server
	fs := http.FileServer(http.Dir(absDir))
	http.Handle("/", fs)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Grimoire Visual Editor Server\n")
	fmt.Printf("Serving directory: %s\n", absDir)
	fmt.Printf("Server running at: http://localhost%s\n", addr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	// Log requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		fs.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// Middleware for logging
// Commented out as it's not currently used
// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
// 		next.ServeHTTP(w, r)
// 	})
// }
