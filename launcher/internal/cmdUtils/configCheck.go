package cmdUtils

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/likexian/doh"
	"github.com/likexian/doh/dns"
	"github.com/luskaner/ageLANServer/common"
	"net/netip"
	"time"
)

func hostToIps(host string, IPv4 bool, IPv6 bool) mapset.Set[netip.Addr] {
	client := doh.Use()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rsp, err := client.Query(ctx, dns.Domain(host), dns.TypeA)
	if err != nil {
		return nil
	}
	validIpAddrs := mapset.NewThreadUnsafeSet[netip.Addr]()
	for _, answer := range rsp.Answer {
		if ipAddr, err := netip.ParseAddr(answer.Data); err == nil {
			if (IPv4 && ipAddr.Is4()) || (IPv6 && ipAddr.Is6()) {
				validIpAddrs.Add(ipAddr)
			}
		}
	}
	return validIpAddrs
}

func InternalExternalDnsMismatch(IPv4 bool, IPv6 bool) mapset.Set[string] {
	hosts := mapset.NewThreadUnsafeSet[string]()
	for _, host := range common.AllHosts() {
		internalMapping := mapset.NewThreadUnsafeSet(
			common.AddrOrIPAddrToIPAddrs(host, false, IPv4, IPv6)...,
		)
		if !hostToIps(host, IPv4, IPv6).Equal(internalMapping) {
			hosts.Add(host)
		}
	}
	return hosts
}
