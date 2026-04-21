package handler

import (
	"log/slog"
	"net/http"
)

// notFound handles 404 Not Found errors.
func notFound(slug string, err error, w http.ResponseWriter, r *http.Request) {
	respondWithError(slug, err, w, r, http.StatusNotFound)
}

// internal handles 500 Internal Server Error errors.
func internal(slug string, err error, w http.ResponseWriter, r *http.Request) {
	respondWithError(slug, err, w, r, http.StatusInternalServerError)
}

// respondWithError logs the error and sends an HTTP error response.
func respondWithError(slug string, err error, w http.ResponseWriter, r *http.Request, status int) {
	slog.Error("http error occurred", "slug", slug, "error", err, "path", r.URL.Path)
	http.Error(w, http.StatusText(status), status)
}
