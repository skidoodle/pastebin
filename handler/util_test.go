package handler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateId(t *testing.T) {
	id, err := generateId()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(id))
	for _, char := range id {
		assert.True(t, strings.ContainsRune(charset, char))
	}
}

func TestHighlight(t *testing.T) {
	goContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	// Test with a specific language extension (.go) -> SHOULD highlight
	highlighted, err := highlight(goContent, "go", "monokai")
	assert.NoError(t, err)
	assert.Contains(t, highlighted, `style="color:#f92672"`, "Should highlight Go code with .go extension")

	// Test with a generic extension (.txt) -> SHOULD auto-detect and highlight
	highlighted, err = highlight(goContent, "txt", "monokai")
	assert.NoError(t, err)
	assert.Contains(t, highlighted, `style="color:#f92672"`, "Should auto-detect Go code with .txt extension")

	// NEW TEST: No extension -> SHOULD NOT highlight
	highlighted, err = highlight(goContent, "", "monokai")
	assert.NoError(t, err)
	assert.NotContains(t, highlighted, `style="color:#f92672"`, "Should NOT highlight Go code when no extension is given")
	assert.Contains(t, highlighted, "package main", "Should still contain the original text")
}
