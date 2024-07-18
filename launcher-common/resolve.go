package launcher_common

import (
	mapset "github.com/deckarep/golang-set/v2"
	"launcher-common/executor"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

var cacheTime = 1 * time.Minute
var failedIpToHosts = make(map[string]time.Time)
var failedHostToIps = make(map[string]time.Time)
var ipToHosts = make(map[string]mapset.Set[string])
var hostToIps = make(map[string]mapset.Set[string])
var pingIpRegexp = regexp.MustCompile(`\[(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\]`)
var netbiosRegexp = regexp.MustCompile(`\s*(.*?)\s*<00>`)

func dnsNameToIps(host string) []string {
	names, err := net.LookupIP(host)
	if err != nil {
		return nil
	}
	namesStr := make([]string, 0)
	for _, name := range names {
		if name.To4() != nil {
			namesStr = append(namesStr, name.String())
		}
	}
	return namesStr
}

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

func netbiosNameToIps(host string) []string {
	result := executor.ExecOptions{File: "ping", ExitCode: true, UseWorkingPath: true, Wait: true, Output: true, Args: []string{"-4", "-n", "1", "-w", "1000", host}}.Exec()
	if !result.Success() {
		return nil
	}
	match := pingIpRegexp.FindStringSubmatch(*result.Output)
	if len(match) == 0 {
		return nil
	}
	return []string{match[1]}
}

func isLocalIp(ip net.IP) bool {
	interfaces, err := net.Interfaces()

	if err != nil {
		return false
	}
	var addrs []net.Addr
	for _, i := range interfaces {
		addrs, err = i.Addrs()
		if err != nil {
			return false
		}

		for _, addr := range addrs {
			var currentIp net.IP
			v, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			currentIp = v.IP
			if currentIp.To4() == nil {
				continue
			}

			if currentIp.Equal(ip) {
				return true
			}
		}
	}

	return false
}

func ipToNetbiosName(ip string) *string {
	result := executor.ExecOptions{File: "nbtstat", UseWorkingPath: true, Wait: true, Output: true, Args: []string{"-a", ip}}.Exec()
	if !result.Success() {
		return nil
	}
	match := netbiosRegexp.FindStringSubmatch(*result.Output)
	if len(match) == 0 {
		parsedIp := net.ParseIP(ip)
		if parsedIp != nil && isLocalIp(parsedIp) {
			hostname, err := os.Hostname()
			if err == nil {
				return &hostname
			}
		}
		return nil
	}
	return &match[1]
}

func HostOrIpToIps(host string) mapset.Set[string] {
	if ip := net.ParseIP(host); ip != nil {
		var ips = mapset.NewSet[string]()
		if ip.To4() != nil {
			ips.Add(ip.String())
		} else {
			hosts := IpToHosts(ip.String())
			for _, h := range hosts.ToSlice() {
				ips = ips.Union(HostOrIpToIps(h))
			}
		}
		return ips
	} else {
		cached, cachedIps := cachedHostToIps(host)
		if cached {
			return cachedIps
		}
		ips := mapset.NewSet[string]()
		ipsFromDns := dnsNameToIps(host)
		if ipsFromDns != nil {
			for _, ipStr := range ipsFromDns {
				ips.Add(ipStr)
				cacheMapping(host, ipStr)
			}
		}
		ipsFromNetbios := netbiosNameToIps(host)
		if ipsFromNetbios != nil {
			for _, ipStr := range ipsFromNetbios {
				ips.Add(ipStr)
				cacheMapping(host, ipStr)
			}
		}
		return ips
	}
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
	hostFromNetbios := ipToNetbiosName(ip)
	if hostFromNetbios != nil {
		hosts.Add(*hostFromNetbios)
		cacheMapping(*hostFromNetbios, ip)
	}
	return hosts
}
