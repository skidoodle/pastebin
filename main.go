package main

import (
	"flag"
	"log/slog"
	"net/http"

	"github.com/csehviktor/pastebin/handler"
	"github.com/csehviktor/pastebin/store"
)

func main() {
	addrPtr := flag.String("addr", ":3000", "port to listen on")
	maxSizePtr := flag.Int("max-size", 32*1024, "maximum size of a paste in bytes")
	flag.Parse()

	addr := *addrPtr
	maxSize := int64(*maxSizePtr)

	mux := http.NewServeMux()
	store := store.NewMemoryStore()
	httpHandler := handler.NewHandler(store, maxSize)

	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/style.css")
	})
	mux.HandleFunc("GET /", httpHandler.HandleHome)
	mux.HandleFunc("POST /", httpHandler.HandleSet)
	mux.HandleFunc("GET /{id}", httpHandler.HandleGet)
	mux.HandleFunc("GET /{id}/{theme}", httpHandler.HandleGet)

	slog.Info("starting http server", "addr", addr, "maxSize", maxSize)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
