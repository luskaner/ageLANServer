package common

import (
	"strings"
	"testing"
)

func TestParseCommandArgsFromSliceJoined(t *testing.T) {
	args, err := ParseCommandArgsFromSlice(
		[]string{"--exe", "{GAME_PATH}", "--flag"},
		map[string]string{"GAME_PATH": "some/path"},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 {
		t.Fatalf("separateFields=false must yield exactly one arg, got %v", args)
	}
	if !strings.Contains(args[0], "some/path") || !strings.Contains(args[0], "--flag") {
		t.Fatalf("arg = %q", args[0])
	}
}

func TestEnhancedViperStringToStringSlice(t *testing.T) {
	got := EnhancedViperStringToStringSlice("value")
	if len(got) != 1 || got[0] != "value" {
		t.Fatalf("got %v", got)
	}
}

func TestUserAgentMentionsProject(t *testing.T) {
	ua := UserAgent()
	if !strings.Contains(ua, Name) {
		t.Fatalf("UserAgent %q missing project name %q", ua, Name)
	}
}

func TestAnnounceHeaderMatchesName(t *testing.T) {
	if AnnounceHeader == "" {
		t.Fatal("AnnounceHeader is empty")
	}
	if AnnounceVersionLatest < AnnounceVersion2 {
		t.Fatalf("AnnounceVersionLatest = %d, must be >= %d", AnnounceVersionLatest, AnnounceVersion2)
	}
}
