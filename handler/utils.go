package handler

import (
	"crypto/rand"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func generateId() (string, error) {
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes), nil
}

func highlight(content, ext, theme string) (string, error) {
	lexer := lexers.Get(ext)
	if lexer == nil {
		lexer = lexers.Analyse(content)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}

	formatter := html.New(
		html.WithLineNumbers(false),
		html.Standalone(false),
		html.TabWidth(4),
		html.WithLineNumbers(true),
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
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
