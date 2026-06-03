package internal

import (
	"encoding/binary"
	"net"
	"time"

	battleServerBroadcast "github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/common/battleServer"
)

func addToSlice(dst []byte, src []byte, srcIndex *int) {
	copy(dst[*srcIndex:], src)
	*srcIndex += len(src)
}

func Broadcast(msg battleServer.BroadcastMessage, port uint16) error {
	if conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4bcast, Port: int(port)}); err != nil {
		return err
	} else {
		data := make([]byte, battleServerBroadcast.MinimumSize+len(msg.Name)-1)
		var index int
		addToSlice(data, battleServerBroadcast.Header, &index)
		var idText []byte
		if idText, err = msg.Id.MarshalText(); err != nil {
			return err
		}
		addToSlice(data, idText, &index)
		binary.LittleEndian.PutUint16(data[index:], uint16(len(msg.Name)))
		index += battleServerBroadcast.PortSize
		addToSlice(data, []byte(msg.Name), &index)
		binary.LittleEndian.PutUint16(data[index:], msg.PublicPort)
		index += battleServerBroadcast.PortSize
		binary.LittleEndian.PutUint16(data[index:], msg.WebsocketPort)
		index += battleServerBroadcast.PortSize
		binary.LittleEndian.PutUint16(data[index:], msg.OutOfBandPort)
		go func() {
			defer func(conn *net.UDPConn) {
				_ = conn.Close()
			}(conn)
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			_, _ = conn.Write(data)
			for range ticker.C {
				if _, err = conn.Write(data); err != nil {
					return
				}
			}
		}()
	}
	return nil
}
