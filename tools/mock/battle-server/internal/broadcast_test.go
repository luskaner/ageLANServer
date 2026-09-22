//go:build windows

package internal

import "testing"

func TestAddToSlice(t *testing.T) {
	dst := make([]byte, 10)
	src := []byte("hello")
	idx := 0
	addToSlice(dst, src, &idx)
	if string(dst[:5]) != "hello" {
		t.Fatalf("got %q, want %q", dst[:5], "hello")
	}
	if idx != 5 {
		t.Fatalf("index = %d, want 5", idx)
	}
	addToSlice(dst, []byte("ab"), &idx)
	if string(dst[5:7]) != "ab" {
		t.Fatalf("got %q, want %q", dst[5:7], "ab")
	}
	if idx != 7 {
		t.Fatalf("index = %d, want 7", idx)
	}
}

func TestAddToSliceEmpty(t *testing.T) {
	dst := make([]byte, 4)
	idx := 0
	addToSlice(dst, []byte{}, &idx)
	if idx != 0 {
		t.Fatalf("index = %d, want 0", idx)
	}
}
