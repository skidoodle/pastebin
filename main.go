package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/skidoodle/pastebin/handler"
	"github.com/skidoodle/pastebin/store"
)

var (
	Version = "devel"
)

type config struct {
	addr    string
	maxSize int64
	dbPath  string
	ttl     time.Duration
}

func parseFlags() *config {
	cfg := &config{}
	flag.StringVar(&cfg.addr, "addr", ":3000", "socket address to bind to")
	flag.Int64Var(&cfg.maxSize, "max-size", 10*1024*1024, "maximum size of a paste in bytes")
	flag.StringVar(&cfg.dbPath, "db-path", "pastebin.db", "path to the database file")
	flag.DurationVar(&cfg.ttl, "ttl", 7*24*time.Hour, "time to live for pastes")
	flag.Parse()
	return cfg
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; worker-src 'self' blob:")
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := parseFlags()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	storage, err := store.NewBoltStore(cfg.dbPath, nil)
	if err != nil {
		slog.Error("failed to initialize store", "error", err)
		os.Exit(1)
	}
	defer storage.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				storage.Cleanup(cfg.ttl)
			case <-ctx.Done():
				return
			}
		}
	}()

	h := handler.NewHandler(storage, cfg.maxSize, "view/templates/*.html")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /static", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	mux.HandleFunc("GET /static/{file...}", func(w http.ResponseWriter, r *http.Request) {
		file := r.PathValue("file")

		if file == "" || strings.HasSuffix(file, "/") {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "view/static/"+file)
	})

	mux.HandleFunc("GET /", h.HandleHome)
	mux.HandleFunc("POST /", h.HandleSet)
	mux.HandleFunc("GET /raw/{id}", h.HandleRaw)
	mux.HandleFunc("GET /{id}", h.HandleGet)

	server := &http.Server{
		Addr:    cfg.addr,
		Handler: securityHeadersMiddleware(mux),
	}

	go func() {
		slog.Info("starting http server", "addr", cfg.addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}
}
