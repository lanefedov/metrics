package main

import (
	"log"
	"net/http"
	"os"

	"github.com/lanefedov/metrics/internal/server"
	"github.com/lanefedov/metrics/internal/storage"
)

func main() {
	cfg, err := loadServerConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	store := storage.NewMemStorage()
	h := server.NewHandler(store)

	log.Printf("listening on %s", cfg.address)
	if err := http.ListenAndServe(cfg.address, h); err != nil {
		log.Fatal(err)
	}
}
