package common

import (
	"reflect"
	"testing"

	commonGame "github.com/luskaner/ageLANServer/common/game"
)

func TestCertDomains(t *testing.T) {
	got := CertDomains()
	want := []string{
		"*.playfabapi.com",
		"aoe-api.reliclink.com",
		"*.worldsedgelink.com",
		"*.ageofempires.com",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CertDomains() = %#v, want %#v", got, want)
	}
}

func TestCertDomainsIncludesSelfSigned(t *testing.T) {
	got := CertDomains()
	for _, d := range SelfSignedCertDomains {
		found := false
		for _, g := range got {
			if g == d {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("CertDomains() = %#v missing self-signed domain %q", got, d)
		}
	}
}

func TestSelfSignedCertGame(t *testing.T) {
	cases := map[string]bool{
		commonGame.AoE1: true,
		commonGame.AoE2: true,
		commonGame.AoE3: true,
		commonGame.AoE4: false,
		commonGame.AoM:  false,
	}
	for game, want := range cases {
		if got := SelfSignedCertGame(game); got != want {
			t.Errorf("SelfSignedCertGame(%q) = %v, want %v", game, got, want)
		}
	}
}
