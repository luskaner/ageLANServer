package ip

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
)

func ResolveIps(ip net.IP) (ips []net.IP, targetIps []net.IP) {
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

			if ip.IsUnspecified() || currentIpv4.Equal(ip) {
				ips = append(ips, currentIpv4.Mask(v.Mask))
				var targetIp net.IP
				if i.Flags&net.FlagBroadcast != 0 {
					targetIp = common.CalculateBroadcastIp(currentIpv4, v.Mask)
				} else {
					targetIp = currentIpv4
				}
				targetIps = append(targetIps, targetIp)
			}
		}
	}

	return
}

func ResolveHosts(hosts []string) []net.IP {
	ipsSet := mapset.NewSet[string]()
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
