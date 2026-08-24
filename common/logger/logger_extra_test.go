package commonLogger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrefixPrintf(t *testing.T) {
	var buf bytes.Buffer
	old := logger
	defer func() { logger = old }()

	logger = nil // logger is nil, PrefixPrintf should not panic
	PrefixPrintf("test", "format %d", 42)

	// Initialize with buffer, then call
	logger = nil
	Initialize(&buf)
	PrefixPrintf("TEST", "hello %s", "world")
	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("PrefixPrintf output = %q, want contain %q", buf.String(), "hello world")
	}
	if !strings.Contains(buf.String(), "|TEST|") {
		t.Errorf("PrefixPrintf output = %q, want contain prefix |TEST|", buf.String())
	}
}

func TestPrefixPrintln(t *testing.T) {
	var buf bytes.Buffer
	old := logger
	defer func() { logger = old }()

	logger = nil
	PrefixPrintln("test", "line1", "line2")

	logger = nil
	Initialize(&buf)
	PrefixPrintln("MYMOD", "a", "b")
	out := buf.String()
	if !strings.Contains(out, "a b") {
		t.Errorf("PrefixPrintln output = %q, want contain %q", out, "a b")
	}
	if !strings.Contains(out, "|MYMOD|") {
		t.Errorf("PrefixPrintln output = %q, want contain prefix |MYMOD|", out)
	}
}

func TestPrefixPrintf_NilLogger(t *testing.T) {
	old := logger
	defer func() { logger = old }()
	logger = nil
	// Should not panic
	PrefixPrintf("test", "format %d", 1)
	PrefixPrintln("test", "data")
	Printf("data")
	Println("data")
}

func TestLogRootDate(t *testing.T) {
	got := LogRootDate("/tmp")
	if !strings.HasPrefix(got, filepath.Join("/tmp", "logs")) {
		t.Errorf("LogRootDate = %q, want prefix %q", got, filepath.Join("/tmp", "logs"))
	}
	// Should contain a date-like suffix
	base := filepath.Base(got)
	if len(base) < 10 {
		t.Errorf("LogRootDate base = %q, looks too short", base)
	}
}

func TestLogRootPrefix(t *testing.T) {
	got := logRootPrefix("/myroot")
	want := filepath.Join("/myroot", "logs")
	if got != want {
		t.Errorf("logRootPrefix = %q, want %q", got, want)
	}
}

func TestRootBuffer_NilRoot(t *testing.T) {
	var r *Root
	called := false
	err := r.Buffer("test", func(w io.Writer) {
		called = true
		if w != nil {
			t.Error("nil Root should pass nil writer")
		}
	})
	if err != nil {
		t.Errorf("Buffer on nil Root should return nil error, got %v", err)
	}
	if !called {
		t.Error("function should be called even for nil Root")
	}
}

func TestRootBuffer_WithData(t *testing.T) {
	err, root := NewFile(t.TempDir(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	err = root.Buffer("output", func(w io.Writer) {
		_, _ = w.Write([]byte("buffered content"))
	})
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root.Folder(), "output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "buffered content" {
		t.Errorf("file content = %q, want %q", string(data), "buffered content")
	}
}

func TestRootBuffer_Empty(t *testing.T) {
	err, root := NewFile(t.TempDir(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	err = root.Buffer("empty", func(w io.Writer) {
		// write nothing
	})
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	_, err = os.Stat(filepath.Join(root.Folder(), "empty.txt"))
	if !os.IsNotExist(err) {
		t.Error("empty buffer should not create file")
	}
}

func TestNewOwnFileLogger(t *testing.T) {
	dir := t.TempDir()
	err := NewOwnFileLogger("testlog", dir, "", true)
	if err != nil {
		t.Fatalf("NewOwnFileLogger: %v", err)
	}
	if FileLogger == nil {
		t.Fatal("FileLogger should be set")
	}
	if file == nil {
		t.Fatal("file should be set")
	}
	if FileLogger.Folder() != dir {
		t.Errorf("FileLogger.Folder() = %q, want %q", FileLogger.Folder(), dir)
	}
	_ = file.Close()
	file = nil
}

func TestNewFile_FinalRootFalse(t *testing.T) {
	root := t.TempDir()
	err, l := NewFile(root, "myGame", false)
	if err != nil {
		t.Fatalf("NewFile(finalRoot=false): %v", err)
	}
	if l == nil {
		t.Fatal("Root should not be nil")
	}
	// The folder should contain logs/myGame/<date>
	folder := l.Folder()
	if !strings.Contains(folder, filepath.Join("logs", "myGame")) {
		t.Errorf("folder = %q, want to contain logs/myGame", folder)
	}
	// Directory should exist
	if _, statErr := os.Stat(folder); os.IsNotExist(statErr) {
		t.Error("directory should have been created")
	}
}

func TestCloseFileLog_NilFile(t *testing.T) {
	oldFile := file
	file = nil
	defer func() { file = oldFile }()
	// Should not panic
	CloseFileLog()
}
