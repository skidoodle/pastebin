package handler

import (
	"crypto/rand"
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// generateId generates a random 10-character alphanumeric string.
func generateId() (string, error) {
	bytes := make([]byte, 10)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes), nil
}

// highlight highlights the given content using the specified file extension and theme.
func highlight(content, ext, theme string) (string, error) {
	var lexer chroma.Lexer

	if ext != "" {
		lexer = lexers.Get(ext)
		if lexer != nil && lexer.Config().Name == "plaintext" {
			analysedLexer := lexers.Analyse(content)
			if analysedLexer != nil {
				lexer = analysedLexer
			}
		}
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := html.New(
		html.WithLineNumbers(true),
		html.TabWidth(4),
	)

	style := styles.Get(theme)
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// render renders the given component to the response writer, handling any errors.
func render(component templ.Component, w http.ResponseWriter, r *http.Request) {
	if err := component.Render(r.Context(), w); err != nil {
		internal("could not render template", err, w, r)
	}
}
