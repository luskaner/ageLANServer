package common

import (
	"context"
	"fmt"
	"net"
	"time"
)

const Tld = "com"
const dotTld = "." + Tld
const linkMainDomainSuffix = "link"
const ApiSubdomainSuffix = "-api"
const SubDomain = "aoe" + ApiSubdomainSuffix
const RelicMainDomain = "relic" + linkMainDomainSuffix
const relicDomain = SubDomain + "." + RelicMainDomain + dotTld
const WorldsEdgeMainDomain = "worldsedge" + linkMainDomainSuffix
const worldsEdge = "." + WorldsEdgeMainDomain
const apiWorldsEdge = ApiSubdomainSuffix + worldsEdge + dotTld
const PlayFabDomain = "playfabapi"
const AgeOfEmpires = "ageofempires"
const ApiAgeOfEmpiresSubdomain = "api"
const CdnAgeOfEmpiresSubdomain = "cdn"
const ApiAgeOfEmpires = ApiAgeOfEmpiresSubdomain + "." + AgeOfEmpires + dotTld
const CdnAgeOfEmpires = CdnAgeOfEmpiresSubdomain + "." + AgeOfEmpires + dotTld
const playFabSuffix = "." + PlayFabDomain + dotTld
const SubDomainAge2Prefix = "pb"
const SubDomainReleasePart = "-live-release"

var SelfSignedCertDomains = []string{relicDomain, "*" + worldsEdge + dotTld, "*." + AgeOfEmpires + dotTld}

var hostsCache = make(map[string][]string)

func CertDomains() []string {
	domains := []string{"*" + playFabSuffix}
	domains = append(domains, SelfSignedCertDomains...)
	return domains
}

func SelfSignedCertGame(game string) bool {
	return game != GameAoM
}

func AllHosts(gameId string) (domains []string) {
	if cache, ok := hostsCache[gameId]; ok {
		return cache
	}
	switch gameId {
	case GameAoE1, GameAoE2, GameAoE3:
		domains = []string{relicDomain, SubDomain + worldsEdge + dotTld}
	case GameAoM:
		domains = []string{"athens-live" + apiWorldsEdge, "C15F9" + playFabSuffix, ApiAgeOfEmpires}
	}
	domains = append(domains, CdnAgeOfEmpires)
	if gameId == GameAoE3 {
		domains = append(domains, ApiAgeOfEmpires)
	}
	domains = append(domains, generateDomains(gameId)...)
	hostsCache[gameId] = domains
	return
}

func domainExists(domain string) bool {
	resolver := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := resolver.LookupIPAddr(ctx, domain)
	return err == nil
}

func generateDomains(gameId string) (domains []string) {
	var prefix string
	var releaseMin int
	switch gameId {
	case GameAoE2:
		prefix = SubDomainAge2Prefix
		releaseMin = 2
	case GameAoM:
		prefix = "andromeda"
		releaseMin = 13
	default:
		return
	}
	generateDomainName := func(release int) string {
		return fmt.Sprintf("%s%s%d%s", prefix, SubDomainReleasePart, release, apiWorldsEdge)
	}
	for release := 1; release <= releaseMin; release++ {
		domains = append(domains, generateDomainName(release))
	}
	for release := releaseMin + 1; ; release++ {
		if domain := generateDomainName(release); domainExists(domain) {
			domains = append(domains, generateDomainName(release))
		} else {
			break
		}
	}
	return
}
