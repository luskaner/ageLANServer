package internal

import (
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"
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
