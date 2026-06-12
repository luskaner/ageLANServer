package internal

import (
	"encoding/binary"
	"net"
	"time"

	battleServerBroadcast "github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"golang.org/x/sys/windows"
)

func addToSlice(dst []byte, src []byte, srcIndex *int) {
	copy(dst[*srcIndex:], src)
	*srcIndex += len(src)
}

func Broadcast(msg battleServer.BroadcastMessage, port uint16) error {
	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_DGRAM, windows.IPPROTO_UDP)
	if err != nil {
		return err
	}
	if err = windows.SetsockoptInt(fd, windows.SOL_SOCKET, windows.SO_BROADCAST, 1); err != nil {
		_ = windows.Closesocket(fd)
		return err
	}
	addr := windows.SockaddrInet4{
		Port: int(port),
		Addr: [4]byte(net.IPv4bcast.To4()),
	}

	data := make([]byte, battleServerBroadcast.MinimumSize+len(msg.Name)-1)
	var index int
	addToSlice(data, battleServerBroadcast.Header, &index)
	var idText []byte
	if idText, err = msg.Id.MarshalText(); err != nil {
		_ = windows.Closesocket(fd)
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
		defer func() {
			_ = windows.Closesocket(fd)
		}()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		_ = windows.Sendto(fd, data, 0, &addr)
		for range ticker.C {
			if err = windows.Sendto(fd, data, 0, &addr); err != nil {
				return
			}
		}
	}()

	return nil
}
