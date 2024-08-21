package ip

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
)

func resolveHost(host string) []net.IP {
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

func ResolveHost(host string) []net.IP {
	ip := net.ParseIP(host)
	if ip == nil {
		return resolveHost(host)
	} else if ip.To4() == nil {
		return nil
	}
	return []net.IP{ip}
}
