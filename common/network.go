package common

import (
	"net"
)

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

func IPv4RunningNetworkInterfaces() (map[*net.Interface][]*net.IPNet, error) {
	interfacesAddresses := make(map[*net.Interface][]*net.IPNet)
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var addrs []net.Addr
	for _, iface := range interfaces {
		if iface.Flags&net.FlagRunning == 0 {
			continue
		}
		addrs, err = iface.Addrs()
		if err != nil {
			continue
		}
		var ifAdded bool
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				if !ifAdded {
					interfacesAddresses[&iface] = make([]*net.IPNet, 0)
					ifAdded = true
				}
				interfacesAddresses[&iface] = append(interfacesAddresses[&iface], ipnet)
			}
		}
	}
	return interfacesAddresses, nil
}
