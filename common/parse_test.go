package common

import (
	"reflect"
	"testing"
)

func TestParseCommandArgsFromSlice_SubstitutesPlaceholders(t *testing.T) {
	args, err := ParseCommandArgsFromSlice(
		[]string{"{exe}", "--flag", "{value}"},
		map[string]string{"exe": "/bin/foo", "value": "bar"},
		true,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"/bin/foo", "--flag", "bar"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestParseCommandArgsFromSlice_NoSeparateFields(t *testing.T) {
	args, err := ParseCommandArgsFromSlice(
		[]string{"one", "two", "three"},
		nil,
		false,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"one two three"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestParseCommandArgsFromSlice_QuotedFields(t *testing.T) {
	args, err := ParseCommandArgsFromSlice(
		[]string{`"a b"`, "c"},
		nil,
		true,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a b", "c"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestParseCommandArgsFromSlice_UnterminatedQuoteErrors(t *testing.T) {
	if _, err := ParseCommandArgsFromSlice([]string{`"unterminated`}, nil, true); err == nil {
		t.Fatalf("expected error for unterminated quote, got nil")
	}
}

func TestEnhancedViperStringToStringSlice(t *testing.T) {
	got := EnhancedViperStringToStringSlice("value")
	want := []string{"value"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
