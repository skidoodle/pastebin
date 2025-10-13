package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/skidoodle/pastebin/store"
	"github.com/skidoodle/pastebin/view"
)

// HttpHandler handles HTTP requests.
type HttpHandler struct {
	store   store.Store
	maxSize int64
}

// NewHandler creates a new HttpHandler.
func NewHandler(store store.Store, maxSize int64) *HttpHandler {
	return &HttpHandler{
		store:   store,
		maxSize: maxSize,
	}
}

// HandleSet handles the creation of a new paste.
func (h *HttpHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	if err := r.ParseForm(); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			badRequest("content too large", err, w, r)
		} else {
			badRequest("invalid form data", err, w, r)
		}
		return
	}

	content := r.FormValue("content")
	if content == "" {
		badRequest("bin cannot be empty", nil, w, r)
		return
	}

	id, err := generateId()
	if err != nil {
		internal("could not generate id", err, w, r)
		return
	}

	if err := h.store.Set(id, content); err != nil {
		internal("could not save bin", err, w, r)
		return
	}

	slog.Info("created bin", "id", id)
	http.Redirect(w, r, "/"+id, http.StatusFound)
}

// HandleGet handles the retrieval of a paste.
func (h *HttpHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var ext string

	if index := strings.LastIndex(id, "."); index > 0 {
		ext = id[index+1:]
		id = id[:index]
	}

	content, exists, err := h.store.Get(id)
	if err != nil {
		internal("could not get bin", err, w, r)
		return
	}
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

// HandleHome handles the home page.
func (h *HttpHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	render(view.BinEditorPage(), w, r)
}
