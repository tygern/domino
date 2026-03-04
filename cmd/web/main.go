package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tygern/domino/internal/web"
)

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	mux := web.Handlers()

	fmt.Fprintf(os.Stderr, "Listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
