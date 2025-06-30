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
	interfaces, err := common.IPv4RunningNetworkInterfaces()
	if err == nil {
		for iff := range interfaces {
			if iff.Flags&net.FlagMulticast != 0 {
				multicastIfs = append(multicastIfs, iff)
			}
		}
	}
	for _, port := range ports {
		for _, ip := range ips {
			addr := &net.UDPAddr{
				IP:   ip.To4(),
				Port: port,
			}
			if conn, err := net.ListenUDP("udp4", addr); err == nil {
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
				connections = append(connections, conn)
			}
		}
	}
	return connections
}

func Announce(ip net.IP, multicastGroup net.IP, port int) bool {
	connections := announcementConnections([]net.IP{ip}, []net.IP{multicastGroup}, []int{port})
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
