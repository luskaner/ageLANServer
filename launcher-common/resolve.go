package launcher_common

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

var cacheTime = 1 * time.Minute
var failedIpAddrToAddrs map[netip.Addr]time.Time
var failedAddrToIpAddrs map[string]time.Time
var ipAddrsToAddrs map[netip.Addr]mapset.Set[string]
var addrToIpAddrs map[string]mapset.Set[netip.Addr]

func init() {
	ClearCache()
}

func addrToDnsName(addr string) mapset.Set[string] {
	names, err := net.LookupAddr(addr)
	if err != nil {
		return nil
	}
	return mapset.NewThreadUnsafeSet(names...)
}

func cachedAddrToIpAddrs(addr string) (bool, mapset.Set[netip.Addr]) {
	var cached bool
	var result mapset.Set[netip.Addr]
	var cachedIpAddrs mapset.Set[netip.Addr]
	hostToLower := strings.ToLower(addr)
	if cachedIpAddrs, cached = addrToIpAddrs[hostToLower]; cached {
		result = cachedIpAddrs
	} else if failedTime, ok := failedAddrToIpAddrs[hostToLower]; ok && time.Since(failedTime) < cacheTime {
		cached = true
	}
	return cached, result
}

func cachedIpAddrToAddrs(ipAddr netip.Addr) (bool, mapset.Set[string]) {
	var cached bool
	var result mapset.Set[string]
	var cachedAddrs mapset.Set[string]
	if cachedAddrs, cached = ipAddrsToAddrs[ipAddr]; cached {
		result = cachedAddrs
	} else if failedTime, ok := failedIpAddrToAddrs[ipAddr]; ok && time.Since(failedTime) < cacheTime {
		cached = true
	}
	return cached, result
}

func CacheMapping(addr string, ipAddr netip.Addr) {
	addrToLower := strings.ToLower(addr)
	if _, exists := addrToIpAddrs[addrToLower]; !exists {
		addrToIpAddrs[addrToLower] = mapset.NewThreadUnsafeSet[netip.Addr]()
	}
	addrToIpAddrs[addrToLower].Add(ipAddr)
	if _, exists := ipAddrsToAddrs[ipAddr]; !exists {
		ipAddrsToAddrs[ipAddr] = mapset.NewThreadUnsafeSet[string]()
	}
	ipAddrsToAddrs[ipAddr].Add(addr)
	if _, exists := failedIpAddrToAddrs[ipAddr]; exists {
		delete(failedIpAddrToAddrs, ipAddr)
	}
	if _, exists := failedAddrToIpAddrs[addrToLower]; exists {
		delete(failedAddrToIpAddrs, addrToLower)
	}
}

func ClearCache() {
	failedIpAddrToAddrs = make(map[netip.Addr]time.Time)
	failedAddrToIpAddrs = make(map[string]time.Time)
	ipAddrsToAddrs = make(map[netip.Addr]mapset.Set[string])
	addrToIpAddrs = make(map[string]mapset.Set[netip.Addr])
}

func filterAddrs(ipAddrs mapset.Set[netip.Addr], IPv4 bool, IPv6 bool) mapset.Set[netip.Addr] {
	filtered := mapset.NewThreadUnsafeSet[netip.Addr]()
	for ipAddr := range ipAddrs.Iter() {
		if (IPv4 && ipAddr.Is4()) || (IPv6 && ipAddr.Is6()) {
			filtered.Add(ipAddr)
		}
	}
	return filtered
}

func AddrToIpAddrs(addr string, IPv4 bool, IPv6 bool) mapset.Set[netip.Addr] {
	ipAddrs := mapset.NewThreadUnsafeSet[netip.Addr]()
	if ipAddr, err := netip.ParseAddr(addr); err == nil {
		isIPv4 := ipAddr.Is4()
		if (IPv4 && isIPv4) || (IPv6 && !isIPv4) {
			if ipAddr.IsUnspecified() {
				ipAddrs = ipAddrs.Union(ResolveUnspecifiedIpAddrs(isIPv4))
			} else {
				ipAddrs.Add(ipAddr)
			}
		}
	} else {
		cached, cachedIpAddrs := cachedAddrToIpAddrs(addr)
		if cached {
			ipAddrs = cachedIpAddrs.Clone()
		} else {
			ipsFromDns := common.AddrOrIPAddrToIPAddrs(addr, false, true, true)
			if len(ipsFromDns) > 0 {
				for _, ipAddr := range ipsFromDns {
					CacheMapping(addr, ipAddr)
				}
			}
		}
	}
	return filterAddrs(ipAddrs, IPv4, IPv6)
}

func ResolveUnspecifiedIpAddrs(IPv4 bool) (addrs mapset.Set[netip.Addr]) {
	addrs = mapset.NewThreadUnsafeSet[netip.Addr]()
	interfaces, err := common.RunningNetworkInterfaces(IPv4, !IPv4, false)
	if err != nil {
		return
	}
	for iff, iffIps := range interfaces {
		for _, n := range iffIps {
			addr, ok := netip.AddrFromSlice(n.IP)
			if !ok {
				continue
			}
			if addr.Is6() && addr.IsLinkLocalMulticast() {
				addr = addr.WithZone(strconv.Itoa(iff.Index))
			}
			addrs.Add(addr)
		}
	}
	return
}

func Matches(addr1 string, addr2 string, IPv4 bool, IPv6 bool) bool {
	addr2IpAddrs := AddrToIpAddrs(addr2, IPv4, IPv6)
	addr1IpAddrs := AddrToIpAddrs(addr1, IPv4, IPv6)
	return !addr2IpAddrs.Intersect(addr1IpAddrs).IsEmpty()
}

func IpAddrToAddrs(ipAddr netip.Addr) mapset.Set[string] {
	cached, cachedHosts := cachedIpAddrToAddrs(ipAddr)
	if cached {
		return cachedHosts
	}
	hosts := mapset.NewThreadUnsafeSet[string]()
	hostsFromDns := addrToDnsName(ipAddr.String())
	if !hostsFromDns.IsEmpty() {
		for hostStr := range hostsFromDns.Iter() {
			hosts.Add(hostStr)
			CacheMapping(hostStr, ipAddr)
		}
	}
	return hosts
}
