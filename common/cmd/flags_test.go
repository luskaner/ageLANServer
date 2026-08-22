package cmd

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func newTestFlagSet(t *testing.T) *pflag.FlagSet {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var (
		s  string
		i  int
		b  bool
		b2 bool
		sl []string
		b64 []byte
	)
	fs.StringVarP(&s, "string", "s", "default-string", "")
	fs.IntVarP(&i, "int", "i", 10, "")
	fs.BoolVarP(&b, "enabled", "e", false, "")
	fs.BoolVarP(&b2, "disabled", "d", true, "")
	fs.StringSliceVar(&sl, "hosts", []string{}, "")
	fs.BytesBase64VarP(&b64, "payload", "p", nil, "")
	return fs
}

func TestFlagSetToArgsOnlyNonDefaults(t *testing.T) {
	fs := newTestFlagSet(t)
	if err := fs.Set("string", "custom"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Set("int", "77"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Parse([]string{"--enabled"}); err != nil {
		t.Fatal(err)
	}

	args := FlagSetToArgs(fs, false)
	got := strings.Join(args, " ")
	if !strings.Contains(got, "--string=custom") {
		t.Fatalf("missing changed string: %q", got)
	}
	if !strings.Contains(got, "--int=77") {
		t.Fatalf("missing changed int: %q", got)
	}
	if !strings.Contains(got, "--enabled") {
		t.Fatalf("missing flipped-on bool: %q", got)
	}
	if strings.Contains(got, "--disabled") {
		t.Fatalf("default-true bool must be skipped: %q", got)
	}
	if strings.Contains(got, "default-string") || strings.Contains(got, "--int=10") {
		t.Fatalf("default values must be skipped: %q", got)
	}
}

func TestFlagSetToArgsExplicitFalseBool(t *testing.T) {
	fs := newTestFlagSet(t)
	if err := fs.Set("disabled", "false"); err != nil {
		t.Fatal(err)
	}
	args := FlagSetToArgs(fs, false)
	found := false
	for _, a := range args {
		if a == "--disabled=false" {
			found = true
		}
	}
	if !found {
		t.Fatalf("explicit false must be emitted as --disabled=false: %v", args)
	}
}

func TestFlagSetToArgsStringSlicePerElement(t *testing.T) {
	fs := newTestFlagSet(t)
	if err := fs.Set("hosts", "a.test,b.test"); err != nil {
		t.Fatal(err)
	}
	args := FlagSetToArgs(fs, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--hosts=a.test") || !strings.Contains(joined, "--hosts=b.test") {
		t.Fatalf("slice must be split into one arg per element: %v", args)
	}
	if strings.Contains(joined, "[") || strings.Contains(joined, "]") {
		t.Fatalf("brackets must not leak into args: %v", args)
	}
}

func TestFlagSetToArgsBase64EmittedVerbatim(t *testing.T) {
	fs := newTestFlagSet(t)
	const payload = "AAECAwQ="
	if err := fs.Set("payload", payload); err != nil {
		t.Fatal(err)
	}
	args := FlagSetToArgs(fs, false)
	want := "--payload=" + payload
	found := false
	for _, a := range args {
		if a == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("base64 value must round-trip verbatim (no re-encoding): %v", args)
	}
}

func TestFlagSetToArgsIncludeNameAndPositional(t *testing.T) {
	fs := pflag.NewFlagSet("mybin", pflag.ContinueOnError)
	if err := fs.Parse([]string{"positional1", "positional2"}); err != nil {
		t.Fatal(err)
	}
	args := FlagSetToArgs(fs, true)
	if len(args) != 3 || args[0] != "mybin" || args[1] != "positional1" || args[2] != "positional2" {
		t.Fatalf("args = %v", args)
	}
}

// Proves the removed Get()-re-encode block was dead code: neither pflag byte
// flag type implements Get(), so the type assertion never matched and the
// block could never fire for bytesHex or bytesBase64.
func TestByteFlagValuesDoNotImplementGet(t *testing.T) {
	for _, tc := range []struct {
		typ    string
		create func(*pflag.FlagSet) pflag.Value
	}{
		{"bytesBase64", func(fs *pflag.FlagSet) pflag.Value {
			var b []byte
			fs.BytesBase64VarP(&b, "payload", "p", nil, "")
			return fs.Lookup("payload").Value
		}},
		{"bytesHex", func(fs *pflag.FlagSet) pflag.Value {
			var b []byte
			fs.BytesHexVarP(&b, "key", "k", nil, "")
			return fs.Lookup("key").Value
		}},
	} {
		fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
		value := tc.create(fs)
		if _, implements := value.(interface{ Get() interface{} }); implements {
			t.Errorf("%s unexpectedly implements Get(): the removed re-encode block would fire on it", tc.typ)
		}
		if _, implements := any(value).(interface{ Get() any }); implements {
			t.Errorf("%s unexpectedly implements Get() any-form", tc.typ)
		}
	}
}

// Documents the mechanism of harm the removed block would have caused if it
// ever fired on a bytesHex-like flag: it emitted base64(bytes) where Set()
// expects hex, while Value.String() round-trips losslessly.
func TestBytesHexRoundTripAndReencodeWouldBreak(t *testing.T) {
	fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
	var payload []byte
	fs.BytesHexVarP(&payload, "key", "k", nil, "")
	payload = []byte{0x0a, 0x1b, 0x2c}
	if err := fs.Set("key", "0a1b2c"); err != nil {
		t.Fatal(err)
	}

	// Current behavior: Value.String() re-parses losslessly.
	valueStr := fs.Lookup("key").Value.String()
	fresh := pflag.NewFlagSet("fresh", pflag.ContinueOnError)
	var again []byte
	fresh.BytesHexVarP(&again, "key", "k", nil, "")
	if err := fresh.Parse([]string{"--key=" + valueStr}); err != nil {
		t.Fatalf("Value.String() output not re-parseable: %v", err)
	}
	if !bytes.Equal(again, payload) {
		t.Fatalf("round trip = %x, want %x", again, payload)
	}

	// Old block behavior: base64 of the decoded bytes instead of hex.
	corrupted := "--key=" + base64.StdEncoding.EncodeToString(payload)
	stale := pflag.NewFlagSet("stale", pflag.ContinueOnError)
	var unused []byte
	stale.BytesHexVarP(&unused, "key", "k", nil, "")
	if err := stale.Parse([]string{corrupted}); err == nil {
		t.Fatal("expected the old re-encoded value to be rejected by a bytesHex flag")
	}
}

func TestFlagSetToArgsRoundTripReparse(t *testing.T) {
	src := newTestFlagSet(t)
	for _, set := range [][2]string{
		{"string", "hello world"},
		{"int", "-5"},
	} {
		if err := src.Set(set[0], set[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := src.Parse([]string{"--enabled", "--hosts=x.y,z.w"}); err != nil {
		t.Fatal(err)
	}

	dst := newTestFlagSet(t)
	if err := dst.Parse(FlagSetToArgs(src, false)); err != nil {
		t.Fatalf("re-parse failed: %v\nargs: %v", err, FlagSetToArgs(src, false))
	}
	if v, _ := dst.GetString("string"); v != "hello world" {
		t.Errorf("string = %q", v)
	}
	if v, _ := dst.GetInt("int"); v != -5 {
		t.Errorf("int = %d", v)
	}
	if v, _ := dst.GetBool("enabled"); !v {
		t.Error("bool not preserved")
	}
	if v, _ := dst.GetStringSlice("hosts"); len(v) != 2 || v[0] != "x.y" || v[1] != "z.w" {
		t.Errorf("slice = %v", v)
	}
}
