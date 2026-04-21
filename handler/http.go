package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/skidoodle/pastebin/store"
)

type HttpHandler struct {
	store     store.Store
	maxSize   int64
	templates *template.Template
}

type ViewData struct {
	Title      string
	Content    string
	ID         string
	IsPreview  bool
	TimeAgo    string
	LineCount  int
	GutterSize int
	Error      string
}

func NewHandler(store store.Store, maxSize int64, tmplPattern string) *HttpHandler {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseGlob(tmplPattern))

	return &HttpHandler{
		store:     store,
		maxSize:   maxSize,
		templates: tmpl,
	}
}

func (h *HttpHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	if err := r.ParseForm(); err != nil {
		data := ViewData{
			Title:     "pastebin",
			IsPreview: false,
		}
		if strings.Contains(err.Error(), "http: request body too large") {
			data.Error = "Content too large"
		} else {
			data.Error = "Invalid form data"
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
			internal("could not render template", err, w, r)
		}
		return
	}

	content := r.FormValue("content")
	if content == "" {
		data := ViewData{
			Title:     "pastebin",
			IsPreview: false,
			Error:     "Bin cannot be empty",
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
			internal("could not render template", err, w, r)
		}
		return
	}

	contentHash := hash(content)
	if id, exists, err := h.store.GetIDByHash(contentHash); err == nil && exists {
		slog.Info("deduplicated bin", "id", id)
		http.Redirect(w, r, "/"+id, http.StatusFound)
		return
	}

	id, err := generateId()
	if err != nil {
		internal("could not generate id", err, w, r)
		return
	}

	if err := h.store.Set(id, contentHash, content); err != nil {
		internal("could not save bin", err, w, r)
		return
	}

	slog.Info("created bin", "id", id)
	http.Redirect(w, r, "/"+id, http.StatusFound)
}

func (h *HttpHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if index := strings.LastIndex(id, "."); index > 0 {
		id = id[:index]
	}

	paste, exists, err := h.store.Get(id)
	if err != nil {
		internal("could not get bin", err, w, r)
		return
	}
	if !exists {
		notFound("bin not found", nil, w, r)
		return
	}

	lineCount := strings.Count(paste.Content, "\n") + 1

	data := ViewData{
		Title:      "bin@" + id,
		Content:    paste.Content,
		ID:         id,
		IsPreview:  true,
		TimeAgo:    TimeAgo(paste.CreatedAt),
		LineCount:  lineCount,
		GutterSize: len(strconv.Itoa(lineCount)),
	}

	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		internal("could not render template", err, w, r)
	}
}

func (h *HttpHandler) HandleRaw(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if index := strings.LastIndex(id, "."); index > 0 {
		id = id[:index]
	}

	paste, exists, err := h.store.Get(id)
	if err != nil {
		internal("could not get bin", err, w, r)
		return
	}
	if !exists {
		notFound("bin not found", nil, w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(paste.Content))
}

func (h *HttpHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	data := ViewData{
		Title:     "pastebin",
		IsPreview: false,
	}
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		internal("could not render template", err, w, r)
	}
}
