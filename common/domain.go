package common

import (
	"fmt"

	game2 "github.com/luskaner/ageLANServer/common/game"
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
	return game != game2.AoE4 && game != game2.AoM
}

func GameHostsDirect(gameId string) (domains []string) {
	switch gameId {
	case game2.AoE4:
		for i := 1; i <= 2; i++ {
			domains = append(domains, fmt.Sprintf("%s%d%s", aoe4SubDomainPrefix, i, apiWorldsEdge))
		}
		fallthrough
	case game2.AoE1, game2.AoE2, game2.AoE3:
		domains = []string{relicDomain, SubDomain + worldsEdge + dotTld}
	case game2.AoM:
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
	case game2.AoM:
		domains = append(domains, "c15f9"+playFabSuffix)
	case game2.AoE4:
		domains = append(domains, "ed603"+playFabSuffix)
	}
	domains = append(domains, CdnAgeOfEmpires, ApiAgeOfEmpires)
	if gameId == game2.AoE4 {
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
	case game2.AoE2:
		prefix = SubDomainAge2Prefix
		releaseMin = 2
		subDomainReleasePart = stdSubDomainReleasePart
	case game2.AoE4:
		prefix = aoe4Marker
		releaseMin = 2
		subDomainReleasePart = "-activerelease"
	case game2.AoM:
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
