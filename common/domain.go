package common

import (
	"context"
	"fmt"
	"net"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
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
const ApiAgeOfEmpiresDomain = "ageofempires"
const ApiAgeOfEmpiresSubdomain = "api"
const ApiAgeOfEmpires = ApiAgeOfEmpiresSubdomain + "." + ApiAgeOfEmpiresDomain + dotTld
const playFabSuffix = "." + PlayFabDomain + dotTld
const SubDomainAge2Prefix = "pb"
const SubDomainReleasePart = "-live-release"
const worldsEdgeDomain = SubDomain + worldsEdge + dotTld

var SelfSignedCertDomains = []string{relicDomain, "*" + worldsEdge + dotTld, "*." + ApiAgeOfEmpiresDomain + dotTld}

var hostsCache = make(map[string][]string)
var domainToGameIds = make(map[string]mapset.Set[string])

func CertDomains() []string {
	domains := []string{"*" + playFabSuffix, ApiAgeOfEmpires}
	domains = append(domains, SelfSignedCertDomains...)
	return domains
}

func CacheAllHosts() {
	for _, gameId := range SupportedGames.ToSlice() {
		AllHosts(gameId)
	}
}

func GameIds(domain string) mapset.Set[string] {
	if gameIds, ok := domainToGameIds[domain]; ok {
		return gameIds
	}
	return mapset.NewThreadUnsafeSet[string]()
}

func AllHosts(gameId string) (domains []string) {
	if cache, ok := hostsCache[gameId]; ok {
		return cache
	}
	switch gameId {
	case GameAoE1, GameAoE2, GameAoE3:
		domains = []string{relicDomain, worldsEdgeDomain}
	case GameAoE4:
		domains = []string{"ED603" + playFabSuffix, worldsEdgeDomain, "dr-activerelease1-api.worldsedgelink.com", "dr-" + ApiAgeOfEmpires}
	case GameAoM:
		domains = []string{"athens-live" + apiWorldsEdge, "C15F9" + playFabSuffix, ApiAgeOfEmpires}
	}
	domains = append(domains, generateDomains(gameId)...)
	hostsCache[gameId] = domains
	for _, domain := range domains {
		if _, ok := domainToGameIds[domain]; !ok {
			domainToGameIds[domain] = mapset.NewThreadUnsafeSet[string]()
		}
		domainToGameIds[domain].Add(gameId)
	}
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
		releaseMin = 12
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
