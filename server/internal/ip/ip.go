package ip

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"net/netip"
)

func ResolveHosts(hosts mapset.Set[string], oneByHostAndType bool, IPv4 bool, IPv6 bool) mapset.Set[netip.Addr] {
	ipAddrs := mapset.NewThreadUnsafeSet[netip.Addr]()
	for host := range hosts.Iter() {
		ipAddr, err := netip.ParseAddr(host)
		if err == nil {
			ipAddrs.Add(ipAddr)
		} else {
			ipAddrs = ipAddrs.Union(
				mapset.NewThreadUnsafeSet[netip.Addr](
					common.AddrOrIPAddrToIPAddrs(host, oneByHostAndType, IPv4, IPv6)...,
				),
			)
		}
	}
	return ipAddrs
}
