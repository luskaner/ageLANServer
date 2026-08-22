package common

import (
	"slices"
	"testing"

	commonGame "github.com/luskaner/ageLANServer/common/game"
)

// Seeds the DNS-probe cache so generateDomains never performs network lookups.
func seedGeneratedDomains(tb testing.TB) {
	tb.Helper()
	for _, g := range []string{
		commonGame.AoE1, commonGame.AoE2, commonGame.AoE3, commonGame.AoE4, commonGame.AoM,
	} {
		generatedDomainsCache[g] = nil
	}
	tb.Cleanup(func() {
		for _, g := range []string{
			commonGame.AoE1, commonGame.AoE2, commonGame.AoE3, commonGame.AoE4, commonGame.AoM,
		} {
			delete(generatedDomainsCache, g)
		}
	})
}

// Regression for the discarded AoE4 domains: the release-specific hosts must
// survive alongside the shared relic/worlds-edge ones.
func TestGameHostsAoE4Union(t *testing.T) {
	seedGeneratedDomains(t)
	domains := GameHosts(commonGame.AoE4, false)

	for _, want := range []string{
		aoe4SubDomainPrefix + "1" + apiWorldsEdge,
		aoe4SubDomainPrefix + "2" + apiWorldsEdge,
		relicDomain,
		SubDomain + worldsEdge + dotTld,
	} {
		if !slices.Contains(domains, want) {
			t.Errorf("AoE4 domains missing %q (got %v)", want, domains)
		}
	}
	if len(domains) != 4 {
		t.Errorf("AoE4 domains = %v, want exactly 4 entries", domains)
	}
}

func TestGameHostsSharedGames(t *testing.T) {
	seedGeneratedDomains(t)
	want := []string{relicDomain, SubDomain + worldsEdge + dotTld}
	for _, g := range []string{commonGame.AoE1, commonGame.AoE2, commonGame.AoE3} {
		got := GameHosts(g, false)
		if !slices.Equal(got[:2], want) {
			t.Errorf("%s: got %v, want prefix %v", g, got[:2], want)
		}
	}
}

func TestGameHostsMacOsExclusiveOnlyAoE2(t *testing.T) {
	seedGeneratedDomains(t)
	with := GameHosts(commonGame.AoE2, true)
	if !slices.Contains(with, aoe2MacWorldsEdgeDomain) {
		t.Errorf("AoE2 macOS-exclusive domain missing: %v", with)
	}
	for _, g := range []string{commonGame.AoE1, commonGame.AoE3, commonGame.AoE4, commonGame.AoM} {
		if slices.Contains(GameHosts(g, true), aoe2MacWorldsEdgeDomain) {
			t.Errorf("%s must not include the macOS-exclusive domain", g)
		}
	}
}

func TestAllHostsAdditions(t *testing.T) {
	seedGeneratedDomains(t)
	aom := AllHosts(commonGame.AoM, false)
	if !slices.Contains(aom, "c15f9"+playFabSuffix) || !slices.Contains(aom, ApiAgeOfEmpires) || !slices.Contains(aom, CdnAgeOfEmpires) {
		t.Errorf("AoM AllHosts incomplete: %v", aom)
	}
	aoe4 := AllHosts(commonGame.AoE4, false)
	if !slices.Contains(aoe4, "ed603"+playFabSuffix) || !slices.Contains(aoe4, Aoe4ApiAgeOfEmpires) {
		t.Errorf("AoE4 AllHosts incomplete: %v", aoe4)
	}
}

func TestSelfSignedCertDomainsIncludedInCertDomains(t *testing.T) {
	for _, d := range SelfSignedCertDomains {
		if !slices.Contains(CertDomains(), d) {
			t.Errorf("CertDomains missing self-signed domain %q", d)
		}
	}
}
