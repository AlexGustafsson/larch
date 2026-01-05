package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexGustafsson/larch/internal/api"
)

func main() {
	endpoint := flag.String("endpoint", "http://localhost:8080", "larch api endpoint")
	flag.Parse()

	client := &api.Client{
		Endpoint: *endpoint,
	}

	server := NewServer(client)

	err := http.ListenAndServe(":8082", server)
	if err != http.ErrServerClosed && err != nil {
		slog.Error("Failed to serve", slog.Any("error", err))
		os.Exit(1)
	}
}
