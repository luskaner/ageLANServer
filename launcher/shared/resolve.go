package shared

import (
	mapset "github.com/deckarep/golang-set/v2"
	"net"
	"os"
	"regexp"
	"shared/executor"
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
	hostToLower := strings.ToLower(host)
	if cachedIps, cached := hostToIps[hostToLower]; cached {
		result = cachedIps
	} else if failedTime, ok := failedHostToIps[hostToLower]; ok && time.Since(failedTime) < cacheTime {
		cached = true
	}
	return cached, result
}

func cachedIpToHosts(ip string) (bool, mapset.Set[string]) {
	var cached bool
	var result mapset.Set[string]
	if cachedHosts, cached := ipToHosts[ip]; cached {
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
	output := executor.RunCustomExecutableOutput("ping", "-4", "-n", "1", "-w", "1000", host)
	if output == nil {
		return nil
	}
	match := pingIpRegexp.FindStringSubmatch(*output)
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

	for _, i := range interfaces {
		addrs, err := i.Addrs()
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
	output := executor.RunCustomExecutableOutput("nbtstat", "-a", ip)
	if output == nil {
		return nil
	}
	match := netbiosRegexp.FindStringSubmatch(*output)
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

func ClearResolveCache() {
	ipToHosts = make(map[string]mapset.Set[string])
	hostToIps = make(map[string]mapset.Set[string])
	failedIpToHosts = make(map[string]time.Time)
	failedHostToIps = make(map[string]time.Time)
}
