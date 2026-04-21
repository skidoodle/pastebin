package handler

import (
	"crypto/rand"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/skidoodle/pastebin/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get(id string) (*store.Paste, bool, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*store.Paste), args.Bool(1), args.Error(2)
}

func (m *MockStore) GetIDByHash(hash string) (string, bool, error) {
	args := m.Called(hash)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockStore) Set(id, hash, content string, metadata map[string]interface{}) error {
	args := m.Called(id, hash, content, metadata)
	return args.Error(0)
}

func (m *MockStore) Del(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestHandleHome(t *testing.T) {
	h := NewHandler(nil, 1024, "../view/templates/*.html")
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h.HandleHome(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleSet(t *testing.T) {
	s := new(MockStore)
	h := NewHandler(s, 1024, "../view/templates/*.html")

	t.Run("Success", func(t *testing.T) {
		content := "new content"
		ch := hash(content)
		s.On("GetIDByHash", ch).Return("", false, nil).Once()
		s.On("Set", mock.Anything, ch, content, mock.Anything).Return(nil).Once()

		form := url.Values{"content": {content}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusFound, rr.Code)
	})

	t.Run("Deduplication", func(t *testing.T) {
		content := "existing content"
		ch := hash(content)
		s.On("GetIDByHash", ch).Return("existingID", true, nil).Once()

		form := url.Values{"content": {content}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusFound, rr.Code)
		assert.Equal(t, "/existingID", rr.Header().Get("Location"))
	})

	t.Run("Empty Content", func(t *testing.T) {
		form := url.Values{"content": {""}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Bin cannot be empty")
	})

	t.Run("Too Large", func(t *testing.T) {
		content := strings.Repeat("a", 2048)
		form := url.Values{"content": {content}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Content too large")
	})

	t.Run("Malformed Form", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("content=%zz"))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid form data")
	})

	t.Run("Store Error", func(t *testing.T) {
		content := "error content"
		ch := hash(content)
		s.On("GetIDByHash", ch).Return("", false, nil).Once()
		s.On("Set", mock.Anything, ch, content, mock.Anything).Return(errors.New("db error")).Once()

		form := url.Values{"content": {content}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Generate ID Error", func(t *testing.T) {
		originalReader := rand.Reader
		defer func() { rand.Reader = originalReader }()
		rand.Reader = errorReader{}

		content := "generate error content"
		ch := hash(content)
		s.On("GetIDByHash", ch).Return("", false, nil).Once()

		form := url.Values{"content": {content}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Malformed Form Data", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("content=%zz"))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid form data")
	})

	t.Run("Too Large Content", func(t *testing.T) {
		content := strings.Repeat("a", 2048)
		form := url.Values{"content": {content}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Content too large")
	})
}

func TestHandleGet(t *testing.T) {
	s := new(MockStore)
	h := NewHandler(s, 1024, "../view/templates/*.html")

	t.Run("Found", func(t *testing.T) {
		id := "testid"
		s.On("Get", id).Return(&store.Paste{Content: "hello", CreatedAt: time.Now()}, true, nil).Once()
		req := httptest.NewRequest("GET", "/"+id, nil)
		req.SetPathValue("id", id)
		rr := httptest.NewRecorder()

		h.HandleGet(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "hello")
	})

	t.Run("Not Found", func(t *testing.T) {
		s.On("Get", "missing").Return(nil, false, nil).Once()
		req := httptest.NewRequest("GET", "/missing", nil)
		req.SetPathValue("id", "missing")
		rr := httptest.NewRecorder()

		h.HandleGet(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Store Error", func(t *testing.T) {
		s.On("Get", "error").Return(nil, false, errors.New("db error")).Once()
		req := httptest.NewRequest("GET", "/error", nil)
		req.SetPathValue("id", "error")
		rr := httptest.NewRecorder()

		h.HandleGet(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("With Extension", func(t *testing.T) {
		id := "testid"
		s.On("Get", id).Return(&store.Paste{Content: "hello", CreatedAt: time.Now()}, true, nil).Once()
		req := httptest.NewRequest("GET", "/"+id+".go", nil)
		req.SetPathValue("id", id+".go")
		rr := httptest.NewRecorder()

		h.HandleGet(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "hello")
	})
}

type FailingResponseWriter struct {
	httptest.ResponseRecorder
}

func (f *FailingResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestHandleHomeError(t *testing.T) {
	h := NewHandler(nil, 1024, "../view/templates/*.html")
	req := httptest.NewRequest("GET", "/", nil)
	rr := &FailingResponseWriter{*httptest.NewRecorder()}
	h.HandleHome(rr, req)
}

type mockTemplateStore struct {
	mock.Mock
}

func (m *mockTemplateStore) ExecuteTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return errors.New("template error")
}

func TestHandleRaw(t *testing.T) {
	s := new(MockStore)
	h := NewHandler(s, 1024, "../view/templates/*.html")

	t.Run("Found", func(t *testing.T) {
		id := "testid"
		content := "raw content"
		s.On("Get", id).Return(&store.Paste{Content: content}, true, nil).Once()
		req := httptest.NewRequest("GET", "/raw/"+id, nil)
		req.SetPathValue("id", id)
		rr := httptest.NewRecorder()

		h.HandleRaw(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, content, rr.Body.String())
		assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
	})

	t.Run("With Extension", func(t *testing.T) {
		id := "testid"
		content := "raw content"
		s.On("Get", id).Return(&store.Paste{Content: content}, true, nil).Once()
		req := httptest.NewRequest("GET", "/raw/"+id+".txt", nil)
		req.SetPathValue("id", id+".txt")
		rr := httptest.NewRecorder()

		h.HandleRaw(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, content, rr.Body.String())
	})

	t.Run("Not Found", func(t *testing.T) {
		s.On("Get", "missing").Return(nil, false, nil).Once()
		req := httptest.NewRequest("GET", "/raw/missing", nil)
		req.SetPathValue("id", "missing")
		rr := httptest.NewRecorder()

		h.HandleRaw(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Store Error", func(t *testing.T) {
		s.On("Get", "error").Return(nil, false, errors.New("db error")).Once()
		req := httptest.NewRequest("GET", "/raw/error", nil)
		req.SetPathValue("id", "error")
		rr := httptest.NewRecorder()

		h.HandleRaw(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestHandleGetTemplateError(t *testing.T) {
	s := new(MockStore)
	h := NewHandler(s, 1024, "../view/templates/*.html")
	id := "testid"
	s.On("Get", id).Return(&store.Paste{Content: "hello", CreatedAt: time.Now()}, true, nil).Once()
	req := httptest.NewRequest("GET", "/"+id, nil)
	req.SetPathValue("id", id)
	rr := &FailingResponseWriter{*httptest.NewRecorder()}

	h.HandleGet(rr, req)
}

func TestHandleSetTemplateError(t *testing.T) {
	h := NewHandler(nil, 1024, "../view/templates/*.html")

	t.Run("Empty Content", func(t *testing.T) {
		form := url.Values{"content": {""}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := &FailingResponseWriter{*httptest.NewRecorder()}
		h.HandleSet(rr, req)
	})

	t.Run("Parse Error", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("content=%zz"))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := &FailingResponseWriter{*httptest.NewRecorder()}
		h.HandleSet(rr, req)
	})
}

func TestTemplateErrors(t *testing.T) {
	tmpl := template.New("empty")
	h := &HttpHandler{
		templates: tmpl,
		maxSize:   1024,
	}

	t.Run("HandleHome Error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		h.HandleHome(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("HandleGet Error", func(t *testing.T) {
		s := new(MockStore)
		h.store = s
		id := "testid"
		s.On("Get", id).Return(&store.Paste{Content: "hello", CreatedAt: time.Now()}, true, nil).Once()
		req := httptest.NewRequest("GET", "/"+id, nil)
		req.SetPathValue("id", id)
		rr := httptest.NewRecorder()
		h.HandleGet(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("HandleSet Empty Content Error", func(t *testing.T) {
		form := url.Values{"content": {""}}
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("HandleSet Parse Error", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("content=%zz"))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.HandleSet(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
func TestSafeHTML(t *testing.T) {
	h := NewHandler(nil, 1024, "../view/templates/*.html")
	tmpl, err := h.templates.New("test").Parse(`{{ safeHTML "<br>" }}`)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	err = tmpl.Execute(rr, nil)
	assert.NoError(t, err)
	assert.Equal(t, "<br>", rr.Body.String())
}
