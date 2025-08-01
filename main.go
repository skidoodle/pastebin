package main

import (
	"flag"
	"log/slog"
	"net/http"

	"github.com/csehviktor/pastebin/handler"
	"github.com/csehviktor/pastebin/store"
)

type cli struct {
	addr    string
	maxSize int64
}

func parse_flags() *cli {
	cli := &cli{}

	flag.StringVar(&cli.addr, "addr", ":3000", "socket address to bind to")
	flag.Int64Var(&cli.maxSize, "max-size", 32*1024, "maximum size of a paste in bytes")

	flag.Parse()

	return cli
}

func main() {
	cli := parse_flags()

	mux := http.NewServeMux()
	store := store.NewMemoryStore()
	httpHandler := handler.NewHandler(store, cli.maxSize)

	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/style.css")
	})
	mux.HandleFunc("GET /", httpHandler.HandleHome)
	mux.HandleFunc("POST /", httpHandler.HandleSet)
	mux.HandleFunc("GET /{id}", httpHandler.HandleGet)
	mux.HandleFunc("GET /{id}/{theme}", httpHandler.HandleGet)

	slog.Info("starting http server", "addr", cli.addr, "maxSize", cli.maxSize)

	err := http.ListenAndServe(cli.addr, mux)
	if err != nil {
		panic(err)
	}
}
