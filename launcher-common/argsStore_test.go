package launcher_common

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestArgsStore_LoadMissingFile(t *testing.T) {
	store := NewArgsStore(filepath.Join(t.TempDir(), "does-not-exist"))
	err, flags := store.Load()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if flags != nil {
		t.Fatalf("expected nil flags for missing file, got %v", flags)
	}
}

func TestArgsStore_StoreThenLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "args")
	store := NewArgsStore(path)
	if err := store.Store([]string{"--a", "--b"}); err != nil {
		t.Fatalf("Store: %v", err)
	}
	err, flags := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Storing into a fresh (empty) file leaves a leading empty entry because the
	// initial empty content splits into a single "" element.
	want := []string{"", "--a", "--b"}
	if !reflect.DeepEqual(flags, want) {
		t.Fatalf("flags = %#v, want %#v", flags, want)
	}
}

func TestArgsStore_StoreDeduplicates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "args")
	store := NewArgsStore(path)
	if err := store.Store([]string{"--a", "--b"}); err != nil {
		t.Fatalf("Store: %v", err)
	}
	if err := store.Store([]string{"--b", "--c"}); err != nil {
		t.Fatalf("Store: %v", err)
	}
	err, flags := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	sort.Strings(flags)
	want := []string{"", "--a", "--b", "--c"}
	if !reflect.DeepEqual(flags, want) {
		t.Fatalf("flags = %#v, want %#v", flags, want)
	}
}

func TestArgsStore_Delete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "args")
	store := NewArgsStore(path)
	if err := store.Store([]string{"--a"}); err != nil {
		t.Fatalf("Store: %v", err)
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat err = %v", err)
	}
	// Deleting again is a no-op.
	if err := store.Delete(); err != nil {
		t.Fatalf("second Delete should be nil, got %v", err)
	}
}
