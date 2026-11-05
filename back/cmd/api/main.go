package main

import (
	"log"
	"net/http"
	"os"

	"contractcheck/back/internal/analysis"
	"contractcheck/back/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	handler := httpapi.NewHandler(analysis.NewService())
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler.Routes()))
}
