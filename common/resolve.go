package common

import (
	"fmt"
	"net"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/miekg/dns"
)

// Google, Cloudfare and OpenDNS primaries then secondaries
var dnsServers = []string{"8.8.8.8", "1.1.1.1", "208.67.222.222", "8.8.4.4", "1.0.0.1", "208.67.220.220"}

var cacheTime = 1 * time.Minute
var failedIpToHosts map[string]time.Time
var failedHostToIps map[string]time.Time
var ipToHosts map[string]mapset.Set[string]
var hostToIps map[string]mapset.Set[string]

func init() {
	ClearCache()
}

func domainToIps(host string) []net.IP {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil
	}
	validIps := make([]net.IP, 0)
	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 != nil {
			validIps = append(validIps, ipv4)
		}
	}
	return validIps
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
	} else if failedTime, ok := failedIpToHosts[ip]; ok && time.Since(failedTime) < cacheTime {
		cached = true
	}
	return cached, result
}

func DirectHostToIP(host string) (string, error) {
	fqdnHost := dns.Fqdn(host)
	m := new(dns.Msg)
	m.SetQuestion(fqdnHost, dns.TypeA)
	client := &dns.Client{
		Timeout: time.Second,
	}
	for _, dnsServer := range dnsServers {
		in, _, err := client.Exchange(m, net.JoinHostPort(dnsServer, "53"))
		if err != nil {
			continue
		}

		if in.Rcode != dns.RcodeSuccess {
			continue
		}

		for _, ans := range in.Answer {
			if a, ok := ans.(*dns.A); ok {
				return a.A.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IP found for %s", host)
}

func CacheMapping(host string, ip string) {
	hostToLower := strings.ToLower(host)
	if _, exists := hostToIps[hostToLower]; !exists {
		hostToIps[hostToLower] = mapset.NewThreadUnsafeSet[string]()
	}
	hostToIps[hostToLower].Add(ip)
	if _, exists := ipToHosts[ip]; !exists {
		ipToHosts[ip] = mapset.NewThreadUnsafeSet[string]()
	}
	ipToHosts[ip].Add(host)
	if _, exists := failedIpToHosts[ip]; exists {
		delete(failedIpToHosts, ip)
	}
	if _, exists := failedHostToIps[hostToLower]; exists {
		delete(failedHostToIps, hostToLower)
	}
}

func ClearCache() {
	failedIpToHosts = make(map[string]time.Time)
	failedHostToIps = make(map[string]time.Time)
	ipToHosts = make(map[string]mapset.Set[string])
	hostToIps = make(map[string]mapset.Set[string])
}

func HostOrIpToIps(host string) []string {
	if ip := net.ParseIP(host); ip != nil {
		var ips []string
		if ip.To4() != nil {
			if ip.IsUnspecified() {
				ips = append(ips, ResolveUnspecifiedIps()...)
			} else {
				ips = append(ips, ip.String())
			}
		}
		return ips
	}

	cached, cachedIps := cachedHostToIps(host)
	if cached {
		return cachedIps.Clone().ToSlice()
	}
	var ips []string
	ipsFromDns := domainToIps(host)
	if ipsFromDns != nil {
		for _, ipRaw := range ipsFromDns {
			ipStr := ipRaw.String()
			ips = append(ips, ipStr)
			CacheMapping(host, ipStr)
		}
	}
	return ips
}

func HostOrIpToIpsSet(host string) mapset.Set[string] {
	return mapset.NewSet[string](HostOrIpToIps(host)...)
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
	addr2Ips := HostOrIpToIpsSet(addr2)
	addr1Ips := HostOrIpToIpsSet(addr1)
	return addr2Ips.Intersect(addr1Ips).Cardinality() > 0
}

func IpToHosts(ip string) mapset.Set[string] {
	cached, cachedHosts := cachedIpToHosts(ip)
	if cached {
		return cachedHosts
	}
	hosts := mapset.NewThreadUnsafeSet[string]()
	hostsFromDns := ipToDnsName(ip)
	if hostsFromDns != nil {
		for _, hostStr := range hostsFromDns {
			hosts.Add(hostStr)
			CacheMapping(hostStr, ip)
		}
	}
	return hosts
}
