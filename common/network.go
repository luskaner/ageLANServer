package common

import (
	"cmp"
	mapset "github.com/deckarep/golang-set/v2"
	"net"
	"net/netip"
	"slices"
)

func sortIPAddrs(ipAddrs mapset.Set[netip.Addr]) (ipAddrsSorted []netip.Addr) {
	ipAddrsSorted = ipAddrs.ToSlice()
	slices.SortFunc(ipAddrsSorted, func(ipA, ipB netip.Addr) int {
		isLoopbackA := ipA.IsLoopback()
		isLoopbackB := ipB.IsLoopback()
		if isLoopbackA && !isLoopbackB {
			return -1
		}
		if !isLoopbackA && isLoopbackB {
			return 1
		}
		isPrivateA := ipA.IsPrivate()

		isPrivateB := ipB.IsPrivate()
		if isPrivateA && !isPrivateB {
			return -1
		}
		if !isPrivateA && isPrivateB {
			return 1
		}
		return cmp.Compare(ipA.String(), ipB.String())
	})
	return ipAddrsSorted
}

func AddrOrIPAddrToIPAddrs(addrStr string, oneByType bool, IPv4 bool, IPv6 bool) (ipAddrs []netip.Addr) {
	ips, err := net.LookupIP(addrStr)
	if err != nil {
		return
	}
	IPv4s := mapset.NewThreadUnsafeSet[netip.Addr]()
	IPv6s := mapset.NewThreadUnsafeSet[netip.Addr]()
	for _, ip := range ips {
		var isIPv4 bool
		if ipv4Addr := NetIPv4ToNetIPAddr(ip); ipv4Addr.IsValid() {
			if IPv4 {
				IPv4s.Add(ipv4Addr)
			}
			isIPv4 = true
		}
		if IPv6 && !isIPv4 {
			if ipv6Addr := NetIPv6ToNetIPAddr(ip); ipv6Addr.IsValid() {
				IPv6s.Add(ipv6Addr)
			}
		}
	}
	if !oneByType {
		return sortIPAddrs(IPv4s.Union(IPv6s))
	}
	if IPv4 && !IPv4s.IsEmpty() {
		ipAddrs = append(ipAddrs, sortIPAddrs(IPv4s)[0])
	}
	if IPv6 && !IPv6s.IsEmpty() {
		var resort bool
		if len(ipAddrs) > 0 {
			resort = true
		}
		ipAddrs = append(ipAddrs, sortIPAddrs(IPv6s)[0])
		if resort {
			ipAddrs = sortIPAddrs(mapset.NewThreadUnsafeSet[netip.Addr](ipAddrs...))
		}
	}
	return
}

func RunningNetworkInterfaces(IPv4 bool, IPv6 bool, includeIPv6LinkLocal bool) (map[*net.Interface][]*net.IPNet, error) {
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
			if ipnet, ok := addr.(*net.IPNet); ok {
				var add bool
				if ipnet.IP.To4() != nil {
					if IPv4 {
						add = true
					}
				} else if IPv6 && ipnet.IP.To16() != nil && (includeIPv6LinkLocal || !ipnet.IP.IsLinkLocalUnicast()) {
					add = true
				}
				if add {
					if !ifAdded {
						interfacesAddresses[&iface] = make([]*net.IPNet, 0)
						ifAdded = true
					}
					interfacesAddresses[&iface] = append(interfacesAddresses[&iface], ipnet)
				}
			}
		}
	}
	return interfacesAddresses, nil
}

func NetIPv4ToNetIPAddr(ip net.IP) (addr netip.Addr) {
	if ipTo4 := ip.To4(); ipTo4 != nil {
		addr = netip.AddrFrom4(([4]byte)(ipTo4))
	}
	return
}

func NetIPv6ToNetIPAddr(ip net.IP) (addr netip.Addr) {
	if ipTo16 := ip.To16(); ipTo16 != nil {
		addr = netip.AddrFrom16(([16]byte)(ipTo16))
	}
	return
}

func NetIPToNetIPAddr(ip net.IP) (addr netip.Addr) {
	if ipv4 := NetIPv4ToNetIPAddr(ip); ipv4.IsValid() {
		return ipv4
	}
	if ipv6 := NetIPv6ToNetIPAddr(ip); ipv6.IsValid() {
		return ipv6
	}
	return
}

func NetIPAddrToNetIP(ip netip.Addr) (addr net.IP) {
	if ip.Is4() {
		ipAs4 := ip.As4()
		addr = ipAs4[:]
	} else if ip.Is6() {
		ipAs16 := ip.As16()
		addr = ipAs16[:]
	}
	return addr
}
