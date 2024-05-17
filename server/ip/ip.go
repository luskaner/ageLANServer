package ip

import "net"

func resolveHost(host string) net.IP {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil
	}
	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 != nil {
			return ipv4
		}
	}
	return nil
}

func calculateBroadcastIp(ip net.IP, mask net.IPMask) net.IP {
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast
}

func ResolveBroadcastIp(ip net.IP) net.IP {
	if ip.IsUnspecified() {
		return net.IPv4bcast
	}

	interfaces, err := net.Interfaces()

	if err != nil {
		return nil
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil
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
				return calculateBroadcastIp(currentIp, v.Mask)
			}
		}
	}

	return nil
}

func ResolveHost(host string) net.IP {
	ip := net.ParseIP(host)
	if ip == nil {
		ip = resolveHost(host)
	} else if ip.To4() == nil {
		return nil
	}
	return ip
}
