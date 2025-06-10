package ip

import (
	"bytes"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"golang.org/x/net/ipv4"
	"net"
)

func announcementConnections(ips []net.IP, multicastGroups []net.IP, ports []int) []*net.UDPConn {
	var connections []*net.UDPConn
	var multicastIfs []*net.Interface
	if len(multicastGroups) > 0 {
		interfaces, err := net.Interfaces()
		if err == nil {
			var addrs []net.Addr
			for _, adapter := range interfaces {
				addrs, err = adapter.Addrs()
				if err != nil {
					continue
				}
				for _, addr := range addrs {
					v, addrOk := addr.(*net.IPNet)
					if !addrOk {
						continue
					}
					var IP net.IP
					if IP = v.IP.To4(); IP == nil {
						continue
					}
					if adapter.Flags&net.FlagRunning != 0 && adapter.Flags&net.FlagMulticast != 0 {
						multicastIfs = append(multicastIfs, &adapter)
					}
				}
			}
		}
	}
	for _, port := range ports {
		for _, ip := range ips {
			addr := &net.UDPAddr{
				IP:   ip,
				Port: port,
			}
			if conn, err := net.ListenUDP("udp4", addr); err == nil {
				if len(multicastGroups) > 0 {
					p := ipv4.NewPacketConn(conn)
					for _, multicastGroup := range multicastGroups {
						multicastAddr := &net.UDPAddr{
							IP:   multicastGroup,
							Port: port,
						}
						for _, multicastIf := range multicastIfs {
							_ = p.JoinGroup(multicastIf, multicastAddr)
						}
					}
				}
				connections = append(connections, conn)
			}
		}
	}
	return connections
}

func Announce(ip net.IP, multicastGroup net.IP, port int) bool {
	var multicastGroups []net.IP
	if multicastGroup != nil {
		multicastGroups = append(multicastGroups, multicastGroup)
	}
	connections := announcementConnections([]net.IP{ip}, multicastGroups, []int{port})
	if len(connections) == 0 {
		return false
	}
	var buf bytes.Buffer
	buf.Write([]byte(common.AnnounceHeader))
	idBuffer, _ := i.Id.MarshalBinary()
	buf.Write(idBuffer)
	write := buf.Bytes()
	for _, conn := range connections {
		go func(conn *net.UDPConn) {
			defer func(conn *net.UDPConn) {
				_ = conn.Close()
			}(conn)
			packetBuffer := make([]byte, len(common.AnnounceHeader))
			for {
				n, clientAddr, err := conn.ReadFromUDP(packetBuffer)
				if err != nil {
					continue
				}
				if n < len(common.AnnounceHeader) || string(packetBuffer) != common.AnnounceHeader {
					continue
				}
				_, _ = conn.WriteToUDP(write, clientAddr)
			}
		}(conn)
	}
	return true
}
