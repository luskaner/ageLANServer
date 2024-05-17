package ip

import (
	"net"
	"time"
)

var data = make([]byte, 1)

func Announce(ip net.IP) {
	data[0] = 43
	broadcastIp := ResolveBroadcastIp(ip)
	udpAddr, err := net.ResolveUDPAddr("udp", broadcastIp.String()+":59999")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		_, _ = conn.Write(data)
	}
}
