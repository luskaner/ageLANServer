package battle_server_broadcast

import (
	"bytes"
	"net"

	"github.com/luskaner/ageLANServer/battle-server-broadcast/internal"
)

var Header = []byte{0x21, 0x24, 0x00}

const GuidLength = 36

const PortSize = 2

var MinimumSize = len(Header) + GuidLength + PortSize + 1 + 3*PortSize

func RetrieveBsInterfaceAddresses() (mostPriority *net.IPNet, restInterfaces []*net.IPNet, err error) {
	var interfaces []net.Interface
	interfaces, err = net.Interfaces()

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
			if internal.FlagsCheck(i.Flags) && internal.FlagsExtraCheck(i.Flags) {
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

func calculateBroadcastIp(ip net.IP, mask net.IPMask) net.IP {
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast
}

func ValidData(data []byte, length int) bool {
	return length >= MinimumSize || bytes.HasPrefix(data, Header)
}

func CloneAnnouncements(mostPriority *net.IPNet, restInterfaces []*net.IPNet, port int) (err error) {
	priorityUdpAddress := &net.UDPAddr{
		IP:   mostPriority.IP,
		Port: port,
	}

	var conn *net.UDPConn
	conn, err = net.ListenUDP("udp", priorityUdpAddress)

	if err != nil {
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	var targets []*net.UDPConn
	for _, restAddress := range restInterfaces {
		var restAddressConn *net.UDPConn
		restAddressConn, err = net.DialUDP(
			"udp4",
			&net.UDPAddr{
				IP: restAddress.IP,
			},
			&net.UDPAddr{
				IP:   calculateBroadcastIp(restAddress.IP.To4(), restAddress.Mask),
				Port: priorityUdpAddress.Port,
			},
		)
		if err == nil {
			targets = append(targets, restAddressConn)
		}
	}

	if len(targets) == 0 {
		return
	}

	defer func() {
		for _, target := range targets {
			_ = target.Close()
		}
	}()

	buffer := make([]byte, 65535)
	var n int
	var addr *net.UDPAddr

	for {
		n, addr, err = conn.ReadFromUDP(buffer)
		if err != nil || !ValidData(buffer, n) || !addr.IP.Equal(mostPriority.IP) {
			continue
		}
		data := buffer[:n]
		for _, target := range targets {
			_, _ = target.Write(data)
		}
	}
}
