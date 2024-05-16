package announce

import (
	"net"
	"time"
)

var data = make([]byte, 1)

func Announce(host string) {
	data[0] = 43
	conn, err := net.Dial("udp", host+":59999")
	if err != nil {
		panic(err)
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		_, err := conn.Write(data)
		if err != nil {
			continue
		}
	}
}
