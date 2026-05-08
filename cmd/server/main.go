package main

import (
	"log"
	"net/http"

	"github.com/lanefedov/metrics/internal/server"
	"github.com/lanefedov/metrics/internal/storage"
)

func main() {
	store := storage.NewMemStorage()
	h := server.NewHandler(store)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", h); err != nil {
		log.Fatal(err)
	}
}
