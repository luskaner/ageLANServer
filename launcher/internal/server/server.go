package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"golang.org/x/net/ipv4"
)

func StartServer(stop string, executable string, args []string, selectBestServerIP func(ips []string) (ok bool, ip string)) (result *commonExecutor.Result, executablePath string, ip string) {
	executablePath = GetExecutablePath(executable)
	if executablePath == "" {
		return
	}
	var showWindow bool
	if stop == "true" {
		showWindow = false
	} else {
		showWindow = true
	}
	result = commonExecutor.Options{File: executablePath, Args: args, ShowWindow: showWindow, Pid: true}.Exec()
	if result.Success() {
		var ok bool
		localIPs := common.HostOrIpToIpsSet(netip.IPv4Unspecified().String()).ToSlice()
		// Wait up to 30s for server to start
		timeout := time.After(30 * time.Second)
	loop:
		for {
			select {
			case <-timeout:
				break loop
			default:
				if ok, ip = selectBestServerIP(localIPs); ok {
					return
				}
			}
		}
		if pid, proc, err := commonProcess.Process(executablePath); err == nil {
			if err = commonProcess.KillPidProc(pid, proc); err != nil {
				fmt.Println("Failed to stop 'server'")
				fmt.Println("Error message: " + err.Error())
				fmt.Println("You may try killing it manually. Kill process 'server' in your task manager.")
			}
		}
		result = nil
	}
	return
}

func GenerateServerCertificates(serverExecutablePath string, canTrustCertificate bool) (errorCode int) {
	if exists, certificateFolder, cert, _, caCert, selfSignedCert, _ := common.CertificatePairs(serverExecutablePath); !exists || CertificateSoonExpired(cert) || CertificateSoonExpired(caCert) || CertificateSoonExpired(selfSignedCert) {
		if !canTrustCertificate {
			fmt.Println("serverStart is true and canTrustCertificate is false. Certificate pair is missing or soon expired. Generate your own certificates manually.")
			errorCode = internal.ErrServerCertMissingExpired
			return
		}
		if certificateFolder == "" {
			fmt.Println("Cannot find certificate folder of the 'server'. Make sure the folder structure of the 'server' is correct.")
			errorCode = internal.ErrServerCertDirectory
			return
		}
		if result := GenerateCertificatePair(certificateFolder); !result.Success() {
			fmt.Println("Failed to generate certificate pair. Check the folder and its permissions")
			errorCode = internal.ErrServerCertCreate
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			return
		}
	}
	return
}

func GetExecutablePath(executable string) string {
	if executable == "auto" {
		return common.FindExecutablePath(common.GetExeFileName(true, common.Server))
	}
	return executable
}

func LanServer(host string, insecureSkipVerify bool) bool {
	ips := common.HostOrIpToIps(host)
	var ip string
	if len(ips) == 0 {
		ip = host
	} else {
		ip = ips[0]
	}
	tr := &http.Transport{
		TLSClientConfig: TlsConfig(host, insecureSkipVerify),
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Head(fmt.Sprintf("https://%s/test", ip))
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func announcementConnections(multicastIPs []net.IP, ports []int) []*net.UDPConn {
	var connections []*net.UDPConn
	var multicastIfs []*net.Interface
	if len(multicastIPs) > 0 {
		interfaces, err := net.Interfaces()
		if err == nil {
			var addrs []net.Addr
			for _, i := range interfaces {
				addrs, err = i.Addrs()
				if err != nil {
					continue
				}
				for _, addr := range addrs {
					v, addrOk := addr.(*net.IPNet)
					if !addrOk {
						continue
					}
					var IP net.IP
					if IP = v.IP.To4(); IP == nil {
						continue
					}
					if i.Flags&net.FlagRunning != 0 && i.Flags&net.FlagMulticast != 0 {
						multicastIfs = append(multicastIfs, &i)
					}
				}
			}
		}
	}
	for _, port := range ports {
		addr := &net.UDPAddr{
			IP:   netip.IPv4Unspecified().AsSlice(),
			Port: port,
		}
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			continue
		}
		if len(multicastIPs) > 0 {
			p := ipv4.NewPacketConn(conn)
			for _, multicastIP := range multicastIPs {
				multicastAddr := &net.UDPAddr{
					IP:   multicastIP,
					Port: port,
				}
				for _, multicastIf := range multicastIfs {
					_ = p.JoinGroup(multicastIf, multicastAddr)
				}
			}
		}
		connections = append(connections, conn)
	}
	return connections
}

func LanServersAnnounced(multicastIPs []net.IP, ports []int) map[uuid.UUID]*common.AnnounceMessage {
	results := make(chan map[uuid.UUID]*common.AnnounceMessage)
	connections := announcementConnections(multicastIPs, ports)
	for _, conn := range connections {
		go func() {
			defer func(conn *net.UDPConn) {
				_ = conn.Close()
			}(conn)

			err := conn.SetReadDeadline(time.Now().Add(15 * time.Second))
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
				_, serverAddr, err = conn.ReadFromUDP(packetBuffer)
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
				case common.AnnounceVersion1:
					var msg common.AnnounceMessageData001
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
						Ips:     mapset.NewThreadUnsafeSet[string](),
					}
					servers[id] = m
				}
				m.Ips.Add(ip)
			}

			results <- servers
		}()
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
