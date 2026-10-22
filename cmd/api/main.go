package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok", "app": "MyInvestor"}`)
	})

	port := ":8080"
	slog.Info("Starting MyInvestor server", "port", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
