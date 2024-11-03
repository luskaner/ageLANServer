package ip

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"net"
)

func ResolveAddrs(listenIP net.IP, multicastIP net.IP, targetPort int, broadcast bool, multicast bool) (sourceIPs []net.IP, targetAddrs []*net.UDPAddr) {
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
			var currentIP net.IP
			v, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			currentIP = v.IP
			currentIPv4 := currentIP.To4()
			if currentIPv4 == nil {
				continue
			}

			if listenIP.IsUnspecified() || currentIPv4.Equal(listenIP) {
				if broadcast {
					sourceIPs = append(sourceIPs, currentIPv4)
					if i.Flags&net.FlagBroadcast != 0 {
						targetAddrs = append(targetAddrs, &net.UDPAddr{
							IP:   common.CalculateBroadcastIp(currentIPv4, v.Mask),
							Port: targetPort,
						})
					} else {
						targetAddrs = append(targetAddrs, &net.UDPAddr{
							IP:   currentIPv4,
							Port: targetPort,
						})
					}
				}

				if multicast && i.Flags&net.FlagMulticast != 0 {
					sourceIPs = append(sourceIPs, currentIPv4)
					targetAddrs = append(
						targetAddrs,
						&net.UDPAddr{
							IP:   multicastIP,
							Port: targetPort,
						},
					)
				}
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
