package battleServer

import (
	"errors"
	"fmt"
	"log"
	"net"
	"slices"
	"strconv"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/process"
)

func StartAndCheck(gameId string) (err error) {
	var path string
	var ok bool
	if ok, path = battleServer.ResolvePath(gameId); !ok {
		return fmt.Errorf("could not find battle server executable")
	}
	// TODO: Add other arguments when implemented
	options := exec.Options{
		File:           path,
		UseWorkingPath: true,
		Pid:            true,
	}
	if result := options.Exec(); result.Success() {
		log.Println("Started battle server with PID:", result.Pid)
		if proc, err := process.FindProcess(int(result.Pid)); err == nil {
			defer func() {
				_ = process.KillProc(proc)
			}()
		}
		var conn *net.UDPConn
		conn, err = net.ListenUDP(
			"udp4",
			&net.UDPAddr{IP: net.IPv4zero, Port: int(battleServer.BroadcastPort(gameId))},
		)
		if err != nil {
			return fmt.Errorf("failed to listen for battle server broadcast: %w", err)
		}
		defer func(conn *net.UDPConn) {
			_ = conn.Close()
		}(conn)

		deadline := time.Now().Add(15 * time.Second)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}
		localIPs := common.ResolveUnspecifiedIps()
		buffer := make([]byte, 65535)
		var msg *battleServer.BroadcastMessage
		var ip net.IP
		log.Println("Listening on battle server broadcast messages...")
		for {
			n, addr, readErr := conn.ReadFromUDP(buffer)
			if readErr != nil {
				if netErr, ok := errors.AsType[net.Error](readErr); ok && netErr.Timeout() {
					break
				}
				log.Printf("Network error: %v\n", readErr)
				break
			}
			if !slices.Contains(localIPs, addr.IP.String()) {
				continue
			}
			if tmpMsg, parseErr := battleServer.ParseBroadcastMessage(buffer, n); parseErr == nil {
				if tmpMsg.Name == battleServer.DefaultName {
					msg = tmpMsg
					ip = addr.IP
					break
				}
			} else {
				log.Printf("Failed to parse broadcast message: %v\n", parseErr)
			}
		}
		if msg == nil {
			return fmt.Errorf("did not receive an appropriate battle server broadcast message within the deadline")
		}
		ports := []uint16{msg.PublicPort, msg.WebsocketPort, msg.OutOfBandPort}
		for _, port := range ports {
			if port == 0 {
				continue
			}
			target := net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
			testConn, portErr := net.DialTimeout("tcp4", target, 100*time.Millisecond)
			if portErr != nil {
				return fmt.Errorf("failed to connect to battle server on port %d: %w", port, portErr)
			}
			_ = testConn.Close()
		}
	} else {
		return fmt.Errorf("could not start battle server executable")
	}
	return nil
}
