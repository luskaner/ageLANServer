package common

const Name = "ageLANServer"
const linkSuffix = "link.com"
const RelicDomain = "aoe-api.relic" + linkSuffix
const worldsEdgeSuffix = "-api.worldsedge" + linkSuffix
const WorldsEdgeDomain1 = "aoe" + worldsEdgeSuffix
const WorldsEdgeDomain2 = "pb-live-release1" + worldsEdgeSuffix
const AthensDomain1 = "athens-live" + worldsEdgeSuffix
const Cert = "cert.pem"
const Key = "key.pem"
const CertSubjectOrganization = "github.com/luskaner/" + Name

func AllHosts() []string {
	return []string{RelicDomain, WorldsEdgeDomain1, WorldsEdgeDomain2, AthensDomain1}
}
