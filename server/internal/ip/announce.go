package ip

import (
	"bytes"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
	"golang.org/x/net/ipv4"
	"net"
	"time"
)

func Announce(listenIP net.IP, multicastIP net.IP, targetBroadcastPort int, broadcast bool, multicast bool) {
	sourceIPs, targetAddrs := ResolveAddrs(listenIP, multicastIP, targetBroadcastPort, broadcast, multicast)
	if len(sourceIPs) == 0 {
		fmt.Println("No suitable addresses found.")
		return
	}
	announce(sourceIPs, targetAddrs)
}

func announce(sourceIPs []net.IP, targetAddrs []*net.UDPAddr) {
	var connections []*net.UDPConn
	for i := range targetAddrs {
		sourceAddr := net.UDPAddr{IP: sourceIPs[i]}
		targetAddr := targetAddrs[i]
		conn, err := net.DialUDP(
			"udp4",
			&sourceAddr,
			targetAddr,
		)
		if targetAddr.IP.IsMulticast() {
			p := ipv4.NewPacketConn(conn)
			_ = p.SetMulticastLoopback(true)
		}
		if err != nil {
			continue
		}
		fmt.Printf("Announcing %s -> %s\n", sourceAddr.IP.String(), targetAddr.IP.String())
		connections = append(connections, conn)
	}

	if len(connections) == 0 {
		fmt.Println("All connections failed.")
		return
	}

	defer func(conns []*net.UDPConn) {
		for _, conn := range conns {
			_ = conn.Close()
		}
	}(connections)

	ticker := time.NewTicker(common.AnnouncePeriod)
	defer ticker.Stop()

	var buf bytes.Buffer
	buf.Write([]byte(common.AnnounceHeader))
	buf.WriteByte(internal.AnnounceVersionLatest)
	uuidBytes, err := internal.Id.MarshalBinary()
	if err != nil {
		fmt.Println("Error generating ID.")
		return
	}
	buf.Write(uuidBytes)
	bufBytes := buf.Bytes()

	for range ticker.C {
		for _, conn := range connections {
			_, _ = conn.Write(bufBytes)
		}
	}
}
