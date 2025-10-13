package main

//go:generate go tool templ generate

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skidoodle/pastebin/handler"
	"github.com/skidoodle/pastebin/store"
)

// config holds the application configuration.
type config struct {
	addr    string
	maxSize int64
	dbPath  string
	ttl     time.Duration
}

// parseFlags parses command-line flags.
func parseFlags() *config {
	cfg := &config{}
	flag.StringVar(&cfg.addr, "addr", ":3000", "socket address to bind to")
	flag.Int64Var(&cfg.maxSize, "max-size", 32*1024, "maximum size of a paste in bytes")
	flag.StringVar(&cfg.dbPath, "db-path", "pastebin.db", "path to the database file")
	flag.DurationVar(&cfg.ttl, "ttl", 7*24*time.Hour, "time to live for pastes (e.g., 24h, 7d)")
	flag.Parse()
	return cfg
}

func main() {
	cfg := parseFlags()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	storage, err := store.NewBoltStore(cfg.dbPath)
	if err != nil {
		slog.Error("failed to initialize store", "error", err)
		os.Exit(1)
	}
	defer storage.Close()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			storage.Cleanup(cfg.ttl)
		}
	}()

	httpHandler := handler.NewHandler(storage, cfg.maxSize)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/style.css")
	})
	mux.HandleFunc("GET /", httpHandler.HandleHome)
	mux.HandleFunc("POST /", httpHandler.HandleSet)

	mux.HandleFunc("GET /{id}", httpHandler.HandleGet)
	mux.HandleFunc("GET /{id}/", httpHandler.HandleGet)
	mux.HandleFunc("GET /{id}/{theme}", httpHandler.HandleGet)
	mux.HandleFunc("GET /{id}/{theme}/", httpHandler.HandleGet)

	server := &http.Server{
		Addr:    cfg.addr,
		Handler: mux,
	}

	go func() {
		slog.Info("starting http server", "addr", cfg.addr, "maxSize", cfg.maxSize, "dbPath", cfg.dbPath, "ttl", cfg.ttl)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "err", err)
		os.Exit(1)
	}

	slog.Info("server exited gracefully")
}
