package internal

import (
	"bytes"
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
)

var header = []byte{0x21, 0x24, 0x00}

const guidLength = 36

const uint16Size = 2

var minimumSize = len(header) + guidLength + uint16Size + 1 + 3*uint16Size

func CloneAnnouncements(mostPriority *net.IPNet, restInterfaces []*net.IPNet) {
	priorityUdpAddress := &net.UDPAddr{
		IP:   mostPriority.IP,
		Port: 9999,
	}

	conn, err := net.ListenUDP("udp", priorityUdpAddress)

	if err != nil {
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	var targets []*net.UDPConn
	for _, restAddress := range restInterfaces {
		var restAddressConn *net.UDPConn
		restAddressConn, err = net.DialUDP(
			"udp",
			&net.UDPAddr{
				IP:   restAddress.IP,
				Port: priorityUdpAddress.Port,
			},
			&net.UDPAddr{
				IP:   common.CalculateBroadcastIp(restAddress.IP.To4(), restAddress.Mask),
				Port: priorityUdpAddress.Port,
			},
		)
		if err == nil {
			targets = append(targets, restAddressConn)
		}
	}

	if len(targets) == 0 {
		return
	}

	defer func() {
		for _, target := range targets {
			_ = target.Close()
		}
	}()

	buffer := make([]byte, 65_535)
	var n int
	var addr *net.UDPAddr

	for {
		n, addr, err = conn.ReadFromUDP(buffer)
		if err != nil || n < minimumSize || !bytes.HasPrefix(buffer, header) || !addr.IP.Equal(mostPriority.IP) {
			continue
		}
		data := buffer[:n]
		for _, target := range targets {
			_, _ = target.Write(data)
		}
	}
}
