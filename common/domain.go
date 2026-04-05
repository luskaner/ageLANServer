package common

import (
	"fmt"
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
const apiAgeOfEmpiresSuffix = "." + AgeOfEmpires + dotTld
const ApiAgeOfEmpires = ApiAgeOfEmpiresSubdomain + apiAgeOfEmpiresSuffix
const Aoe4ApiAgeOfEmpires = ApiAgeOfEmpiresSubdomain + "-" + aoe4Marker + apiAgeOfEmpiresSuffix
const CdnAgeOfEmpires = CdnAgeOfEmpiresSubdomain + "." + AgeOfEmpires + dotTld
const playFabSuffix = "." + PlayFabDomain + dotTld
const SubDomainAge2Prefix = "pb"
const stdSubDomainReleasePart = "-live-release"
const aoe4SubDomainPrefix = "aoeliverelease"
const aoe4Marker = "dr"

var SelfSignedCertDomains = []string{relicDomain, "*" + worldsEdge + dotTld, "*." + AgeOfEmpires + dotTld}

var hostsCache = make(map[string][]string)

func CertDomains() []string {
	domains := []string{"*" + playFabSuffix}
	domains = append(domains, SelfSignedCertDomains...)
	return domains
}

func SelfSignedCertGame(game string) bool {
	return game != GameAoE4 && game != GameAoM
}

func GameHostsDirect(gameId string) (domains []string) {
	switch gameId {
	case GameAoE4:
		for i := 1; i <= 2; i++ {
			domains = append(domains, fmt.Sprintf("%s%d%s", aoe4SubDomainPrefix, i, apiWorldsEdge))
		}
		fallthrough
	case GameAoE1, GameAoE2, GameAoE3:
		domains = []string{relicDomain, SubDomain + worldsEdge + dotTld}
	case GameAoM:
		domains = []string{"athens-live" + apiWorldsEdge}
	}
	domains = append(domains, generateDomains(gameId)...)
	return domains
}

func AllHosts(gameId string) (domains []string) {
	if cache, ok := hostsCache[gameId]; ok {
		return cache
	}
	domains = GameHostsDirect(gameId)
	switch gameId {
	case GameAoM:
		domains = append(domains, "c15f9"+playFabSuffix)
	case GameAoE4:
		domains = append(domains, "ed603"+playFabSuffix)
	}
	domains = append(domains, CdnAgeOfEmpires, ApiAgeOfEmpires)
	if gameId == GameAoE4 {
		domains = append(domains, Aoe4ApiAgeOfEmpires)
	}
	hostsCache[gameId] = domains
	return
}

func generateDomains(gameId string) (domains []string) {
	var prefix string
	var releaseMin int
	var subDomainReleasePart string
	switch gameId {
	case GameAoE2:
		prefix = SubDomainAge2Prefix
		releaseMin = 2
		subDomainReleasePart = stdSubDomainReleasePart
	case GameAoE4:
		prefix = aoe4Marker
		releaseMin = 2
		subDomainReleasePart = "-activerelease"
	case GameAoM:
		prefix = "andromeda"
		releaseMin = 15
		subDomainReleasePart = stdSubDomainReleasePart
	default:
		return
	}
	generateDomainName := func(release int) string {
		return fmt.Sprintf("%s%s%d%s", prefix, subDomainReleasePart, release, apiWorldsEdge)
	}
	for release := 1; release <= releaseMin; release++ {
		domains = append(domains, generateDomainName(release))
	}
	for release := releaseMin + 1; ; release++ {
		if _, err := DirectHostToIP(generateDomainName(release)); err == nil {
			domains = append(domains, generateDomainName(release))
		} else {
			break
		}
	}
	return
}
