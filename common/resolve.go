package common

import "net"

func HostToIps(host string) []net.IP {
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
