package store

import "testing"

func TestMemoryStoreGet(t *testing.T) {
	store := NewMemoryStore()

	store.Set("key", "value")
	if val, _ := store.Get("key"); val != "value" {
		t.Errorf("get() = %s, want %s", val, "value")
	}
}

func TestMemoryStoreExists(t *testing.T) {
	store := NewMemoryStore()

	if _, exists := store.Get("something"); exists {
		t.Errorf("get() = %t, want %t", exists, false)
	}
}

func TestMemoryStoreOverride(t *testing.T) {
	store := NewMemoryStore()

	store.Set("key", "value")
	store.Set("key", "new_value")
	if val, _ := store.Get("key"); val != "new_value" {
		t.Errorf("get() = %s, want %s", val, "new_value")
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	store := NewMemoryStore()

	store.Set("key", "value")
	if _, exists := store.Get("key"); !exists {
		t.Errorf("get() = %t, want %t", exists, true)
	}

	store.Del("key")
	if val, _ := store.Get("key"); val != "" {
		t.Errorf("del() = %s, want %s", val, "")
	}
}
