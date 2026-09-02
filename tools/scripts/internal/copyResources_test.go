package internal

import (
	"path/filepath"
	"testing"
)

func TestBuildResourcePath(t *testing.T) {
	got := BuildResourcePath("server")
	want := filepath.Join("build", "server", "resources")
	if got != want {
		t.Fatalf("BuildResourcePath(\"server\") = %q, want %q", got, want)
	}
}

func TestResourcePath(t *testing.T) {
	got := ResourcePath("launcher")
	want := filepath.Join("launcher", "resources")
	if got != want {
		t.Fatalf("ResourcePath(\"launcher\") = %q, want %q", got, want)
	}
}
