package handler

import (
	"log/slog"
	"net/http"
)

func notFound(slug string, err error, w http.ResponseWriter, r *http.Request) {
	respondWithError(slug, err, w, r, http.StatusNotFound)
}

func badRequest(slug string, err error, w http.ResponseWriter, r *http.Request) {
	respondWithError(slug, err, w, r, http.StatusInternalServerError)
}

func internal(slug string, err error, w http.ResponseWriter, r *http.Request) {
	respondWithError(slug, err, w, r, http.StatusInternalServerError)
}

func respondWithError(slug string, err error, w http.ResponseWriter, r *http.Request, status int) {
	slog.Error("http error occured", "slug", slug, "error", err, "path", r.URL.Path)
	http.Error(w, slug, status)
}
