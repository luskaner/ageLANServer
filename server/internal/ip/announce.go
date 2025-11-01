package ip

import (
	"bytes"
	"net"
	"net/netip"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"golang.org/x/net/ipv4"
)

func QueryConnections(ipAddr netip.Addr, multicastGroups mapset.Set[netip.Addr], port int) (err error, conns []*net.UDPConn) {
	var multicastIfs []*net.Interface
	var interfaces map[*net.Interface][]*net.IPNet
	interfaces, err = common.RunningNetworkInterfaces()
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
	addr := &net.UDPAddr{
		IP:   common.NetIPAddrToNetIP(ipAddr),
		Port: port,
	}
	var conn *net.UDPConn
	if conn, err = net.ListenUDP("udp4", addr); err == nil {
		var pckConn *ipv4.PacketConn
		if !multicastGroups.IsEmpty() {
			pckConn = ipv4.NewPacketConn(conn)
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
