package cmdUtils

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/luskaner/ageLANServer/common/logger"
)

func GeneratePortsAsNeeded(ports []int) (generatedPorts []int, err error) {
	var missingIndexes []int
	for i, port := range ports {
		if port == 0 {
			missingIndexes = append(missingIndexes, i)
		}
	}
	missingIndexesCount := len(missingIndexes)
	if missingIndexesCount > 0 {
		commonLogger.Println("Generating ports...")
		generatedPorts, err = findUnusedPorts(missingIndexesCount)
		if err != nil {
			return nil, err
		}
		for _, i := range missingIndexes {
			missingIndexesCount--
			ports[i] = generatedPorts[missingIndexesCount]
		}
	}
	return ports, nil
}

func Available(port int) bool {
	address := net.JoinHostPort(netip.IPv4Unspecified().String(), fmt.Sprintf("%d", port))
	err, listener := listenTCP(address)
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

func listenTCP(address string) (err error, listener net.Listener) {
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func findUnusedPorts(count int) ([]int, error) {
	var ports []int
	var listeners []net.Listener
	address := net.JoinHostPort(netip.IPv4Unspecified().String(), "0")
	var err error
	for i := 0; i < count; i++ {
		var listener net.Listener
		err, listener = listenTCP(address)
		if err != nil {
			break
		}
		listeners = append(listeners, listener)
		ports = append(ports, listener.Addr().(*net.TCPAddr).Port)
	}

	for _, l := range listeners {
		_ = l.Close()
	}

	if err != nil {
		return nil, err
	}

	return ports, nil
}
