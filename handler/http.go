package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/csehviktor/pastebin/store"
	"github.com/csehviktor/pastebin/view"
)

type HttpHandler struct {
	store   *store.MemoryStore
	maxSize int64
}

func NewHandler(store *store.MemoryStore, maxSize int64) *HttpHandler {
	return &HttpHandler{
		store,
		maxSize,
	}
}

func (h *HttpHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	// form body request looks like:
	// content=...
	// so +8 additional bytes must be included
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize+8)

	if err := r.ParseForm(); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			badRequest("content too large", nil, w, r)
		} else {
			badRequest("invalid form data", err, w, r)
		}
		return
	}

	content := r.FormValue("content")
	if content == "" {
		badRequest("bin cant be empty", nil, w, r)
		return
	}

	//fmt.Println(len(content))

	id, err := generateId()
	if err != nil {
		internal("could not generate id", err, w, r)
		return
	}

	h.store.Set(id, content)
	slog.Info("created bin", "id", id)
	http.Redirect(w, r, "/"+id, http.StatusFound)
}

func (h *HttpHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ext := "txt"

	if index := strings.Index(id, "."); index > 0 {
		if index <= len(id) {
			ext = id[index+1:]
			id = id[:index]
		}
	}

	content, exists := h.store.Get(id)
	if !exists {
		notFound("bin not found", nil, w, r)
		return
	}

	theme := r.PathValue("theme")
	if theme == "" {
		theme = "catppuccin-macchiato"
	}

	highlighted, err := highlight(content, ext, theme)
	if err != nil {
		internal("could not highlight content", err, w, r)
		return
	}

	render(view.BinPreviewPage(id, highlighted), w, r)
}

func (h *HttpHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	render(view.BinEditorPage(), w, r)
}
