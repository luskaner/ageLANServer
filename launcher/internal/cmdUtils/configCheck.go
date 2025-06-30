package cmdUtils

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/likexian/doh"
	"github.com/likexian/doh/dns"
	"github.com/luskaner/ageLANServer/common"
	"net"
	"time"
)

func hostToIps(host string) []net.IP {
	client := doh.Use()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rsp, err := client.Query(ctx, dns.Domain(host), dns.TypeA)
	if err != nil {
		return nil
	}
	validIps := make([]net.IP, 0)
	for _, answer := range rsp.Answer {
		if ip := net.ParseIP(answer.Data); ip != nil {
			if ipv4 := ip.To4(); ipv4 != nil {
				validIps = append(validIps, ip)
			}
		}
	}
	return validIps
}

func InternalExternalDnsMismatch() mapset.Set[string] {
	hosts := mapset.NewThreadUnsafeSet[string]()
	ipSliceToStringSet := func(ips []net.IP) mapset.Set[string] {
		set := mapset.NewThreadUnsafeSetWithSize[string](len(ips))
		for _, ip := range ips {
			set.Add(ip.String())
		}
		return set
	}
	for _, host := range common.AllHosts() {
		externalMapping := ipSliceToStringSet(hostToIps(host))
		internalMapping := ipSliceToStringSet(common.HostToIps(host))
		if !internalMapping.Equal(externalMapping) {
			hosts.Add(host)
		}
	}
	return hosts
}
