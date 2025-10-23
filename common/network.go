package common

import (
	"net"
	"net/netip"

	mapset "github.com/deckarep/golang-set/v2"
)

func RunningNetworkInterfaces() (map[*net.Interface][]*net.IPNet, error) {
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
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					if _, ok := interfacesAddresses[&iface]; !ok {
						interfacesAddresses[&iface] = make([]*net.IPNet, 0)
					}
					interfacesAddresses[&iface] = append(interfacesAddresses[&iface], ipnet)
				}
			}
		}
	}
	return interfacesAddresses, nil
}

func NetIPToNetIPAddr(ip net.IP) (addr netip.Addr) {
	if ipTo4 := ip.To4(); ipTo4 != nil {
		addr = netip.AddrFrom4(([4]byte)(ipTo4))
	}
	return
}

func NetIPAddrToNetIP(ip netip.Addr) (addr net.IP) {
	ipAs4 := ip.As4()
	addr = ipAs4[:]
	return addr
}

func NetIPSliceToNetIPSet(ips []net.IP) mapset.Set[netip.Addr] {
	ipAddrs := mapset.NewThreadUnsafeSetWithSize[netip.Addr](len(ips))
	for _, ip := range ips {
		ipAddrs.Add(NetIPToNetIPAddr(ip))
	}
	return ipAddrs
}

func StringSliceToNetIPSlice(ips []string) []net.IP {
	netIPs := make([]net.IP, 0)
	for _, ipStr := range ips {
		if ip := net.ParseIP(ipStr); ip != nil {
			netIPs = append(netIPs, ip)
		}
	}
	return netIPs
}
