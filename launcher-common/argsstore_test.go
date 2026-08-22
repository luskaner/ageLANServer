package launcher_common

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func newTempStore(t *testing.T) (*ArgsStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "args.txt")
	return NewArgsStore(path), path
}

func TestArgsStoreLoadMissingIsClean(t *testing.T) {
	store, _ := newTempStore(t)
	err, flags := store.Load()
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(flags) != 0 {
		t.Fatalf("flags = %v, want empty", flags)
	}
}

// Regression: a zero-byte file (created but crashed before writing) produced a
// phantom [""] flag, which downstream made ConfigRevert attempt a full
// all-games revert instead of a no-op.
func TestArgsStoreLoadEmptyFileYieldsNoFlags(t *testing.T) {
	store, path := newTempStore(t)
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	err, flags := store.Load()
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(flags) != 0 {
		t.Fatalf("empty file produced phantom flags %v", flags)
	}
}

func TestArgsStoreStoreLoadRoundTripAndDedup(t *testing.T) {
	store, _ := newTempStore(t)

	if err := store.Store([]string{"--a", "--b"}); err != nil {
		t.Fatal(err)
	}
	// Storing again with overlap must not duplicate.
	if err := store.Store([]string{"--b", "--c"}); err != nil {
		t.Fatal(err)
	}

	err, flags := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"--a": true, "--b": true, "--c": true}
	if len(flags) != len(want) {
		t.Fatalf("flags = %v, want exactly %v", flags, want)
	}
	for _, f := range flags {
		if !want[f] {
			t.Fatalf("unexpected flag %q in %v", f, flags)
		}
	}
}

func TestArgsStoreDeleteIdempotent(t *testing.T) {
	store, path := newTempStore(t)
	if err := store.Store([]string{"--x"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("first Delete: %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatal("file not removed")
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("second Delete must be nil, got %v", err)
	}
}

func TestArgsStoreConcurrentSameInstance(t *testing.T) {
	store, _ := newTempStore(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			flag := "--flag" + string(rune('a'+i))
			_ = store.Store([]string{flag})
			_, _ = store.Load()
		}(i)
	}
	wg.Wait()
	err, flags := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(flags) < 1 {
		t.Fatal("expected at least one stored flag")
	}
}
