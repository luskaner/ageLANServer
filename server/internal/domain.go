package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"golang.org/x/net/publicsuffix"
)

var age2Pb = regexp.MustCompile(
	fmt.Sprintf(
		`^%s%s\d+%s$`,
		common.SubDomainAge2Prefix,
		common.SubDomainReleasePart,
		common.ApiSubdomainSuffix,
	),
)

func SplitDomain(domain string) (subdomain, mainDomain, tld string, err error) {
	lowerDomain := strings.ToLower(domain)
	var etldPlusOne string
	etldPlusOne, err = publicsuffix.EffectiveTLDPlusOne(lowerDomain)
	if err != nil {
		return
	}
	parts := strings.SplitN(etldPlusOne, ".", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid domain: %s", domain)
		return
	}
	mainDomain = parts[0]
	tld = parts[1]
	subdomain = ""
	if strings.HasSuffix(lowerDomain, "."+etldPlusOne) {
		subdomain = strings.TrimSuffix(lowerDomain, "."+etldPlusOne)
	}
	return
}

func SelfSignedCertificate(domain string) bool {
	subdomain, mainDomain, tld, err := SplitDomain(domain)
	if err != nil || tld != common.Tld {
		return false
	}
	return (subdomain == common.SubDomain && mainDomain == common.RelicMainDomain) ||
		(subdomain == common.SubDomain && mainDomain == common.WorldsEdgeMainDomain) ||
		(age2Pb.MatchString(subdomain) && mainDomain == common.WorldsEdgeMainDomain)
}
