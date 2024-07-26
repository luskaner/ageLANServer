package server

import (
	"bytes"
	"common"
	"crypto/tls"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"golang.org/x/sys/windows"
	launcherCommon "launcher-common"
	commonExecutor "launcher-common/executor"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"time"
)

var autoServerDir = []string{`\`, `\..\`, fmt.Sprintf(`\..\%s\`, common.Server)}
var autoServerName = []string{common.GetExeFileName(common.Server)}

func StartServer(stop string, executable string, args []string) (result *commonExecutor.ExecResult, ip string) {
	executablePath := GetExecutablePath(executable)
	if executablePath == "" {
		return
	}
	var windowState int
	if stop == "true" {
		windowState = windows.SW_HIDE
	} else {
		windowState = windows.SW_MINIMIZE
	}
	result = commonExecutor.ExecOptions{File: executablePath, Args: args, WindowState: windowState, Pid: true}.Exec()
	if result.Success() {
		// Wait up to 30s for server to start
		for i := 0; i < 30; i++ {
			for curIp := range launcherCommon.HostOrIpToIps(netip.IPv4Unspecified().String()).Iter() {
				if LanServer(curIp, true) {
					ip = curIp
					return
				}
			}
			time.Sleep(time.Second)
		}
		if err := commonExecutor.Kill(int(result.Pid)); err != nil {
			fmt.Println("Failed to stop server")
			fmt.Println("Error message: " + result.Err.Error())
		}
		result = nil
	}
	return
}

func GetExecutablePath(executable string) string {
	if executable == "auto" {
		ex, err := os.Executable()
		if err != nil {
			return ""
		}
		exePath := filepath.Dir(ex)
		var f os.FileInfo
		for _, dir := range autoServerDir {
			dirPath := exePath + dir
			for _, name := range autoServerName {
				p := dirPath + name
				if f, err = os.Stat(p); err == nil && !f.IsDir() {
					return path.Clean(p)
				}
			}
		}
		return ""
	}
	return executable
}

func LanServer(host string, insecureSkipVerify bool) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://" + host + "/test")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func LanServersAnnounced(ports []int) map[uuid.UUID]*common.AnnounceMessage {
	results := make(chan map[uuid.UUID]*common.AnnounceMessage)

	for _, port := range ports {
		go func(port int) {
			addr := net.UDPAddr{
				IP:   netip.IPv4Unspecified().AsSlice(),
				Port: port,
			}
			conn, err := net.ListenUDP("udp", &addr)
			if err != nil {
				return
			}
			defer func(conn *net.UDPConn) {
				_ = conn.Close()
			}(conn)

			err = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
			if err != nil {
				return
			}

			packetBuffer := make([]byte, 65_536)
			headerBuffer := make([]byte, len(common.AnnounceHeader))
			var messageLenBuffer uint16
			var messageBuffer *bytes.Buffer
			servers := make(map[uuid.UUID]*common.AnnounceMessage)
			var n int
			var serverAddr *net.UDPAddr

			for {
				n, serverAddr, err = conn.ReadFromUDP(packetBuffer)
				if err != nil {
					break
				}
				n = copy(headerBuffer, packetBuffer)
				if n < len(common.AnnounceHeader) || string(headerBuffer) != common.AnnounceHeader {
					continue
				}
				remainingPacketBuffer := packetBuffer[n:]
				version := remainingPacketBuffer[:common.AnnounceVersionLength][0]
				remainingPacketBuffer = remainingPacketBuffer[common.AnnounceVersionLength:]
				var id uuid.UUID
				id, err = uuid.FromBytes(remainingPacketBuffer[:common.AnnounceIdLength])
				if err != nil {
					continue
				}
				remainingPacketBuffer = remainingPacketBuffer[common.AnnounceIdLength:]
				err = binary.Read(bytes.NewReader(remainingPacketBuffer[2:]), binary.LittleEndian, &messageLenBuffer)
				if err != nil {
					continue
				}
				remainingPacketBuffer = remainingPacketBuffer[2:]
				messageBuffer = bytes.NewBuffer(remainingPacketBuffer[:messageLenBuffer])
				var data interface{}
				switch version {
				case common.AnnounceVersion0:
					var msg common.AnnounceMessageData000
					dec := gob.NewDecoder(messageBuffer)
					if err = dec.Decode(&msg); err == nil {
						data = msg
					}
				}
				ip := serverAddr.IP.String()
				var m *common.AnnounceMessage
				var ok bool
				if m, ok = servers[id]; !ok {
					m = &common.AnnounceMessage{
						Version: version,
						Data:    data,
						Ips:     mapset.NewSet[string](),
					}
					servers[id] = m
				}
				m.Ips.Add(ip)
			}

			results <- servers
		}(port)
	}

	servers := make(map[uuid.UUID]*common.AnnounceMessage)
	for range ports {
		for id, server := range <-results {
			if _, ok := servers[id]; !ok {
				servers[id] = server
			} else {
				for ip := range server.Ips.Iter() {
					servers[id].Ips.Add(ip)
				}
			}
		}
	}

	return servers
}
