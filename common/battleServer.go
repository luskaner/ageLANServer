package common

import "net"

func RetrieveBsInterfaceAddresses() (mostPriority *net.IPNet, restInterfaces []*net.IPNet) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return
	}

	var addrs []net.Addr
	for _, i := range interfaces {
		addrs, err = i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ipNet *net.IPNet
			if ipnet, ok := addr.(*net.IPNet); ok {
				ipNet = ipnet
			} else {
				continue
			}

			if ipNet.IP.To4() == nil {
				continue
			}
			if i.Flags&net.FlagUp != 0 && i.Flags&net.FlagLoopback == 0 &&
				i.Flags&net.FlagRunning != 0 && i.Flags&net.FlagBroadcast != 0 {
				if mostPriority == nil {
					mostPriority = ipNet
				} else {
					restInterfaces = append(restInterfaces, ipNet)
				}
			}
		}
	}
	return
}
