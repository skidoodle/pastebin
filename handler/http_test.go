package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of the store.Store interface.
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get(key string) (string, bool, error) {
	args := m.Called(key)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockStore) Set(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockStore) Del(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// TestHandleHome tests the HandleHome method of the Handler.
func TestHandleHome(t *testing.T) {
	h := NewHandler(nil, 1024)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	h.HandleHome(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestHandleSet tests the HandleSet method of the Handler.
func TestHandleSet(t *testing.T) {
	mockStore := new(MockStore)
	h := NewHandler(mockStore, 1024)

	// Test successful creation
	mockStore.On("Set", mock.Anything, "test content").Return(nil).Once()
	form := url.Values{}
	form.Add("content", "test content")
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	h.HandleSet(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	mockStore.AssertExpectations(t)

	// Test empty content
	req = httptest.NewRequest("POST", "/", strings.NewReader("content="))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	h.HandleSet(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleGet(t *testing.T) {
	mockStore := new(MockStore)
	h := NewHandler(mockStore, 1024)

	// Test found
	mockStore.On("Get", "testid").Return("test content", true, nil).Once()
	req := httptest.NewRequest("GET", "/testid", nil)
	req.SetPathValue("id", "testid")
	rr := httptest.NewRecorder()

	h.HandleGet(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "test content")
	mockStore.AssertExpectations(t)

	// Test not found
	mockStore.On("Get", "notfound").Return("", false, nil).Once()
	req = httptest.NewRequest("GET", "/notfound", nil)
	req.SetPathValue("id", "notfound")
	rr = httptest.NewRecorder()

	h.HandleGet(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockStore.AssertExpectations(t)
}
