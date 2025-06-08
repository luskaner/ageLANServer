package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"golang.org/x/net/ipv4"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

var autoServerDir = []string{fmt.Sprintf(`%c%s%c`, filepath.Separator, common.Server, filepath.Separator), fmt.Sprintf(`%c..%c`, filepath.Separator, filepath.Separator), fmt.Sprintf(`%c..%c%s%c`, filepath.Separator, filepath.Separator, common.Server, filepath.Separator)}
var autoServerName = []string{common.GetExeFileName(true, common.Server)}

type MesuredIpAddress struct {
	Ip      string
	Latency time.Duration
}

func StartServer(gameId string, id uuid.UUID, stop string, executablePath string, args []string) (result *commonExecutor.Result, ip string) {
	result = commonExecutor.Options{File: executablePath, Args: args, ShowWindow: stop != "true", Pid: true}.Exec()
	if result.Success() {
		localIPs := launcherCommon.HostOrIpToIps(netip.IPv4Unspecified().String())
		// Wait up to 30s for server to start
		timeout := time.After(30 * time.Second)
	loop:
		for {
			select {
			case <-timeout:
				break loop
			default:
				if ips, data := FilterServerIPs(id, gameId, localIPs); data != nil {
					ip = ips[0].Ip
					return
				}
				if _, err := commonProcess.FindProcess(result.Pid); err != nil {
					break loop
				}
			}
		}
		if proc, err := commonProcess.FindProcess(result.Pid); err == nil {
			fmt.Print(
				printer.Gen(
					printer.Error,
					"",
					printer.T("Could not connect after starting it, stopping it... "),
				),
			)
			if err = commonProcess.KillProc("", proc); err != nil {
				printer.PrintFailedError(err)
			} else {
				printer.PrintSucceeded()
			}
		}
		result = nil
	}
	return
}

func ResolveExecutablePath() string {
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

func LanServerHost(id uuid.UUID, gameId string, host string, insecureSkipVerify bool) (ok bool) {
	ips := launcherCommon.HostOrIpToIps(host)
	if ips.IsEmpty() {
		return
	}
	for ip := range ips.Iter() {
		if ok, _, _ = LanServerIP(id, gameId, ip, host, insecureSkipVerify); !ok {
			return
		}
	}
	return true
}

func LanServerIP(id uuid.UUID, gameId string, ip string, serverName string, insecureSkipVerify bool) (ok bool, latency time.Duration, data *AnnounceMessageDataSupportedLatest) {
	if id == uuid.Nil {
		return
	}
	tr := &http.Transport{
		TLSClientConfig: TlsConfig(serverName, insecureSkipVerify),
	}
	client := &http.Client{Transport: tr}
	start := time.Now()
	resp, err := client.Get(fmt.Sprintf("https://%s/test", ip))
	latency = time.Since(start)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return
	}
	certificates := resp.TLS.PeerCertificates
	if len(certificates) == 0 {
		return
	}
	// Check if the server is v1.7.3 or higher (minimum version supported)
	if certificates[0].Subject.CommonName != common.Name {
		return
	}
	version := resp.Header.Get(common.VersionHeader)
	serverId := resp.Header.Get(common.IdHeader)
	// Check if the server is v1.7.3 - v1.8.2
	if version == "" && serverId == "" {
		ok = true
		return
	}
	versionInt, _ := strconv.Atoi(version)
	if versionInt > AnnounceVersionSupportedLatest {
		return
	}
	var serverIdUuid uuid.UUID
	serverIdUuid, err = uuid.Parse(serverId)
	if err != nil {
		return
	}
	if id != serverIdUuid {
		return
	}
	data = &AnnounceMessageDataSupportedLatest{}
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return
	}
	if !slices.Contains(data.GameIds, gameId) {
		return
	}
	ok = true
	return
}

func FilterServerIPs(id uuid.UUID, gameId string, possibleIps mapset.Set[string]) (measuredIpAddresses []MesuredIpAddress, data *AnnounceMessageDataSupportedLatest) {
	for curIp := range possibleIps.Iter() {
		if ok, latency, tmpData := LanServerIP(id, gameId, curIp, curIp, true); ok {
			measuredIpAddresses = append(measuredIpAddresses, MesuredIpAddress{
				Ip:      curIp,
				Latency: latency,
			})
			if data == nil {
				data = tmpData
			}
		}
	}
	slices.SortStableFunc(measuredIpAddresses, func(a, b MesuredIpAddress) int {
		return int(a.Latency - b.Latency)
	})
	return
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

func decodeMessage[T any](buff *bytes.Buffer) T {
	var msg T
	dec := gob.NewDecoder(buff)
	_ = dec.Decode(&msg)
	return msg
}

func LanServersAnnounced(ctx context.Context, multicastIPs []net.IP, ports []int) map[uuid.UUID]*AnnounceMessage {
	results := make(chan map[uuid.UUID]*AnnounceMessage)
	connections := announcementConnections(multicastIPs, ports)
	for _, conn := range connections {
		go func() {
			defer func(conn *net.UDPConn) {
				_ = conn.Close()
			}(conn)

			packetBuffer := make([]byte, 65_536)
			headerBuffer := make([]byte, len(common.AnnounceHeader))
			var messageLenBuffer uint16
			var messageBuffer *bytes.Buffer
			servers := make(map[uuid.UUID]*AnnounceMessage)
			var n int
			var serverAddr *net.UDPAddr
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				default:
					err := conn.SetReadDeadline(time.Now().Add(1 * time.Second))
					if err != nil {
						return
					}
					_, serverAddr, err = conn.ReadFromUDP(packetBuffer)

					if err != nil {
						var netErr net.Error
						if errors.As(err, &netErr) && netErr.Timeout() {
							continue
						}
					}

					n = copy(headerBuffer, packetBuffer)
					if n < len(common.AnnounceHeader) || string(headerBuffer) != common.AnnounceHeader {
						continue
					}
					remainingPacketBuffer := packetBuffer[n:]
					version := remainingPacketBuffer[:AnnounceVersionLength][0]
					remainingPacketBuffer = remainingPacketBuffer[AnnounceVersionLength:]
					var id uuid.UUID
					id, err = uuid.FromBytes(remainingPacketBuffer[:AnnounceIdLength])
					if err != nil {
						continue
					}
					var data interface{}
					if version < common.AnnounceVersion2 {
						remainingPacketBuffer = remainingPacketBuffer[AnnounceIdLength:]
						err = binary.Read(bytes.NewReader(remainingPacketBuffer[2:]), binary.LittleEndian, &messageLenBuffer)
						if err != nil {
							continue
						}
						remainingPacketBuffer = remainingPacketBuffer[2:]
						messageBuffer = bytes.NewBuffer(remainingPacketBuffer[:messageLenBuffer])

						switch version {
						case common.AnnounceVersion0:
							data = decodeMessage[common.AnnounceMessageData000](messageBuffer)
						case common.AnnounceVersion1:
							data = decodeMessage[common.AnnounceMessageData001](messageBuffer)
						default:
							data = nil
						}
					}
					ip := serverAddr.IP.String()
					var m *AnnounceMessage
					var ok bool
					if m, ok = servers[id]; !ok {
						m = &AnnounceMessage{
							Version: version,
							Data:    data,
							Ips:     mapset.NewThreadUnsafeSet[string](),
						}
						servers[id] = m
					}
					m.Ips.Add(ip)
				}
			}

			results <- servers
		}()
	}

	servers := make(map[uuid.UUID]*AnnounceMessage)
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
