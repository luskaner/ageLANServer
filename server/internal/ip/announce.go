package ip

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
	"time"
)

func Announce(ip net.IP, port int) {
	ips, targetIps := ResolveIps(ip)

	if len(ips) == 0 {
		fmt.Println("No suitable addresses found.")
		return
	}

	var connections []*net.UDPConn
	for _, targetIp := range targetIps {
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   targetIp,
			Port: port,
		})
		if err != nil {
			continue
		}
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
	err := enc.Encode(common.AnnounceMessageData000{})
	if err != nil {
		fmt.Println("Error encoding message data.")
		return
	}
	messageBuffBytes := messageBuff.Bytes()
	var buf bytes.Buffer
	buf.Write([]byte(common.AnnounceHeader))
	buf.WriteByte(common.AnnounceVersion0)
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
