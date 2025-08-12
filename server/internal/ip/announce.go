package ip

import (
	"bytes"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"net/netip"
	"slices"
)

type packetConn struct {
	ipv4 *ipv4.PacketConn
	ipv6 *ipv6.PacketConn
}

func packetConnFromNetwork(network string, conn *net.UDPConn) *packetConn {
	if network == "udp4" {
		return &packetConn{ipv4: ipv4.NewPacketConn(conn)}
	} else if network == "udp6" {
		return &packetConn{ipv6: ipv6.NewPacketConn(conn)}
	}
	return nil
}

func (p *packetConn) JoinGroup(ifi *net.Interface, group *net.UDPAddr) error {
	if p.ipv4 != nil {
		return p.ipv4.JoinGroup(ifi, group)
	} else if p.ipv6 != nil {
		return p.ipv6.JoinGroup(ifi, group)
	}
	return nil
}

func QueryConnections(ipAddr netip.Addr, multicastGroups mapset.Set[netip.Addr], port int, IPv4 bool, dualStack bool) (err error, conns []*net.UDPConn) {
	var multicastIfs []*net.Interface
	var interfaces map[*net.Interface][]*net.IPNet
	interfaces, err = common.RunningNetworkInterfaces(IPv4, !IPv4, true)
	if err != nil {
		return
	}
	hasUnspecified := ipAddr.IsUnspecified()
	for iff, nets := range interfaces {
		if iff.Flags&net.FlagMulticast == 0 {
			continue
		}
		if hasUnspecified || slices.ContainsFunc(nets, func(ipNet *net.IPNet) bool {
			parsedIPAddr, _ := netip.AddrFromSlice(ipNet.IP)
			return parsedIPAddr == ipAddr
		}) {
			multicastIfs = append(multicastIfs, iff)
		}
	}
	network := "udp"
	if !dualStack {
		if IPv4 {
			network += "4"
		} else {
			network += "6"
		}
	}
	addr := &net.UDPAddr{
		IP:   ipAddr.AsSlice(),
		Port: port,
	}
	var conn *net.UDPConn
	if conn, err = net.ListenUDP(network, addr); err == nil {
		var pckConn *packetConn
		if !multicastGroups.IsEmpty() {
			pckConn = packetConnFromNetwork(network, conn)
		}
		for multicastGroup := range multicastGroups.Iter() {
			multicastAddr := &net.UDPAddr{
				IP:   multicastGroup.AsSlice(),
				Port: port,
			}
			for _, multicastIf := range multicastIfs {
				if err = pckConn.JoinGroup(multicastIf, multicastAddr); err != nil {
					continue
				}
			}
		}
		conns = append(conns, conn)
	}
	return
}

func ListenQueryConnections(connections []*net.UDPConn) {
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
}
