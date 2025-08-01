package main

import (
	"log/slog"
	"net/http"

	"github.com/csehviktor/pastebin/handler"
	"github.com/csehviktor/pastebin/store"
	"github.com/csehviktor/pastebin/view"
)

const addr = ":3000"

func main() {
	mux := http.NewServeMux()
	store := store.NewMemoryStore()
	httpHandler := handler.NewHandler(store)

	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/style.css")
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		handler.Render(view.BinEditorPage(), w, r)
	})
	mux.HandleFunc("POST /", httpHandler.HandleSet)
	mux.HandleFunc("GET /{id}", httpHandler.HandleGet)

	slog.Info("starting http server on", "addr", addr)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
