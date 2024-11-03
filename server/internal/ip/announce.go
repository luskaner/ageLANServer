package ip

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/viper"
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

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var messageBuff bytes.Buffer
	enc := gob.NewEncoder(&messageBuff)
	err := enc.Encode(common.AnnounceMessageData001{
		GameIds: viper.GetStringSlice("Games"),
	})
	if err != nil {
		fmt.Println("Error encoding message data.")
		return
	}
	messageBuffBytes := messageBuff.Bytes()
	var buf bytes.Buffer
	buf.Write([]byte(common.AnnounceHeader))
	buf.WriteByte(common.AnnounceVersion1)
	var uuidBytes []byte
	uuidBytes, err = uuid.New().MarshalBinary()
	if err != nil {
		fmt.Println("Error generating ID.")
		return
	}
	buf.Write(uuidBytes)
	err = binary.Write(&buf, binary.LittleEndian, uint16(len(messageBuffBytes)))
	if err != nil {
		fmt.Println("Error encoding message length.")
		return
	}
	buf.Write(messageBuffBytes)
	bufBytes := buf.Bytes()

	for range ticker.C {
		for _, conn := range connections {
			_, _ = conn.Write(bufBytes)
		}
	}
}
