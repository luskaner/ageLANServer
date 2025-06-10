package ip

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"net"
)

func ResolveHosts(hosts []string) []net.IP {
	ipsSet := mapset.NewThreadUnsafeSet[string]()
	for _, host := range hosts {
		ip := net.ParseIP(host)
		if ip == nil {
			for _, resolvedIP := range common.HostToIps(host) {
				ipsSet.Add(resolvedIP.String())
			}
		} else if ip.To4() != nil {
			ipsSet.Add(ip.String())
		}
	}
	var ips []net.IP
	for _, ip := range ipsSet.ToSlice() {
		ips = append(ips, net.ParseIP(ip))
	}
	return ips
}
