package launcher_common

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
	"strings"
	"time"
)

var cacheTime = 1 * time.Minute
var failedIpToHosts = make(map[string]time.Time)
var failedHostToIps = make(map[string]time.Time)
var ipToHosts = make(map[string]mapset.Set[string])
var hostToIps = make(map[string]mapset.Set[string])

func ipToDnsName(ip string) []string {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return nil
	}
	return names
}

func cachedHostToIps(host string) (bool, mapset.Set[string]) {
	var cached bool
	var result mapset.Set[string]
	var cachedIps mapset.Set[string]
	hostToLower := strings.ToLower(host)
	if cachedIps, cached = hostToIps[hostToLower]; cached {
		result = cachedIps
	} else if failedTime, ok := failedHostToIps[hostToLower]; ok && time.Since(failedTime) < cacheTime {
		cached = true
	}
	return cached, result
}

func cachedIpToHosts(ip string) (bool, mapset.Set[string]) {
	var cached bool
	var result mapset.Set[string]
	var cachedHosts mapset.Set[string]
	if cachedHosts, cached = ipToHosts[ip]; cached {
		result = cachedHosts
	} else if failedTime, ok := failedHostToIps[ip]; ok && time.Since(failedTime) < cacheTime {
		cached = true
	}
	return cached, result
}

func cacheMapping(host string, ip string) {
	hostToLower := strings.ToLower(host)
	if _, exists := hostToIps[hostToLower]; !exists {
		hostToIps[hostToLower] = mapset.NewSet[string]()
	}
	hostToIps[hostToLower].Add(ip)
	if _, exists := ipToHosts[ip]; !exists {
		ipToHosts[ip] = mapset.NewSet[string]()
	}
	ipToHosts[ip].Add(host)
	if _, exists := failedIpToHosts[ip]; exists {
		delete(failedIpToHosts, ip)
	}
	if _, exists := failedIpToHosts[hostToLower]; exists {
		delete(failedHostToIps, hostToLower)
	}
}

func HostOrIpToIps(host string) mapset.Set[string] {
	if ip := net.ParseIP(host); ip != nil {
		var ips = mapset.NewSet[string]()
		if ip.To4() != nil {
			if ip.IsUnspecified() {
				ips.Append(ResolveUnspecifiedIps()...)
			} else {
				ips.Add(ip.String())
			}
		}
		return ips
	} else {
		cached, cachedIps := cachedHostToIps(host)
		if cached {
			return cachedIps
		}
		ips := mapset.NewSet[string]()
		ipsFromDns := common.HostToIps(host)
		if ipsFromDns != nil {
			for _, ipRaw := range ipsFromDns {
				ipStr := ipRaw.String()
				ips.Add(ipStr)
				cacheMapping(host, ipStr)
			}
		}
		return ips
	}
}

func ResolveUnspecifiedIps() (ips []string) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return
	}

	var addrs []net.Addr
	for _, i := range interfaces {

		if i.Flags&net.FlagRunning == 0 {
			continue
		}

		addrs, err = i.Addrs()
		if err != nil {
			return
		}

		for _, addr := range addrs {
			var currentIp net.IP
			v, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			currentIp = v.IP
			currentIpv4 := currentIp.To4()
			if currentIpv4 == nil {
				continue
			}

			ips = append(ips, currentIpv4.String())
		}
	}

	return
}

func Matches(addr1 string, addr2 string) bool {
	addr2Ips := HostOrIpToIps(addr2)
	addr1Ips := HostOrIpToIps(addr1)
	return addr2Ips.Intersect(addr1Ips).Cardinality() > 0
}

func IpToHosts(ip string) mapset.Set[string] {
	cached, cachedHosts := cachedIpToHosts(ip)
	if cached {
		return cachedHosts
	}
	hosts := mapset.NewSet[string]()
	hostsFromDns := ipToDnsName(ip)
	if hostsFromDns != nil {
		for _, hostStr := range hostsFromDns {
			hosts.Add(hostStr)
			cacheMapping(hostStr, ip)
		}
	}
	return hosts
}
