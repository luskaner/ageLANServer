package server

import (
	"context"
	"encoding/json"
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
	"sync"
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
		timeout := time.After(time.Duration(localIPs.Cardinality()) * 3 * time.Second)
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
	version := resp.Header.Get(common.VersionHeader)
	serverId := resp.Header.Get(common.IdHeader)
	if version == "" || serverId == "" {
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

func QueryServers(ctx context.Context, multicastIPs []net.IP, ports []int, broadcast bool, servers map[uuid.UUID]*AnnounceMessage) {
	sourceToTargetAddrs := sourceToTargetUDPAddrs(multicastIPs, ports, broadcast)
	if len(sourceToTargetAddrs) == 0 {
		return
	}
	type connTarget struct {
		conn   *net.UDPConn
		target *net.UDPAddr
	}
	var connTargets []*connTarget
	for source, targets := range sourceToTargetAddrs {
		conn, err := net.ListenUDP(
			"udp4",
			source,
		)
		if err != nil {
			continue
		}
		if slices.ContainsFunc(targets, func(addr *net.UDPAddr) bool {
			return addr.IP.IsMulticast()
		}) {
			p := ipv4.NewPacketConn(conn)
			if err = p.SetMulticastLoopback(true); err != nil {
				continue
			}
		}
		for _, target := range targets {
			connTargets = append(connTargets, &connTarget{
				conn:   conn,
				target: target,
			})
		}
	}

	if len(connTargets) == 0 {
		return
	}

	defer func(connectionPairs []*connTarget) {
		for _, connPair := range connectionPairs {
			_ = connPair.conn.Close()
		}
	}(connTargets)

	data := []byte(common.AnnounceHeader)
	var serverLock sync.Mutex

	sendAndReceive := func(packetBuffer *[]byte, conn *connTarget, servers map[uuid.UUID]*AnnounceMessage) {
		if _, err := conn.conn.WriteToUDP(data, conn.target); err != nil {
			return
		}
		if err := conn.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return
		}
		n, addr, err := conn.conn.ReadFromUDP(*packetBuffer)
		if err != nil {
			return
		}
		if n < len(*packetBuffer) {
			return
		}
		if string((*packetBuffer)[:len(common.AnnounceHeader)]) != common.AnnounceHeader {
			return
		}
		var parsedId uuid.UUID
		parsedId, err = uuid.FromBytes((*packetBuffer)[len(common.AnnounceHeader):])
		if err != nil {
			return
		}
		func() {
			serverLock.Lock()
			defer serverLock.Unlock()
			var server *AnnounceMessage
			var ok bool
			if server, ok = servers[parsedId]; !ok {
				server = &AnnounceMessage{
					Ips: mapset.NewThreadUnsafeSet[string](),
				}
				servers[parsedId] = server
			}
			server.Ips.Add(addr.IP.String())
		}()
	}

	for _, conn := range connTargets {
		go func(conn *connTarget) {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			packetBuffer := make([]byte, len(common.AnnounceHeader)+AnnounceIdLength)
			sendAndReceive(&packetBuffer, conn, servers)
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				case <-ticker.C:
					sendAndReceive(&packetBuffer, conn, servers)
				}
			}
		}(conn)
	}
	<-ctx.Done()
}

func sourceToTargetUDPAddrs(multicastGroups []net.IP, targetPorts []int, broadcast bool) (mapping map[*net.UDPAddr][]*net.UDPAddr) {
	mapping = make(map[*net.UDPAddr][]*net.UDPAddr)
	interfaces, err := net.Interfaces()

	if err != nil {
		return
	}

	var addrs []net.Addr
	for _, i := range interfaces {

		if i.Flags&net.FlagRunning == 0 {
			continue
		}

		addrs, err = i.Addrs()
		if err != nil {
			return
		}

		for _, addr := range addrs {
			var currentIP net.IP
			v, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			currentIP = v.IP
			currentIPv4 := currentIP.To4()
			if currentIPv4 == nil {
				continue
			}
			sourceAddr := &net.UDPAddr{
				IP: currentIPv4,
			}
			mapping[sourceAddr] = make([]*net.UDPAddr, 0)
			if broadcast && i.Flags&net.FlagBroadcast != 0 {
				for _, port := range targetPorts {
					mapping[sourceAddr] = append(mapping[sourceAddr], &net.UDPAddr{
						IP:   common.CalculateBroadcastIp(currentIPv4, v.Mask),
						Port: port,
					})
				}
			} else {
				for _, port := range targetPorts {
					mapping[sourceAddr] = append(mapping[sourceAddr], &net.UDPAddr{
						IP:   currentIPv4,
						Port: port,
					})
				}
			}

			if len(multicastGroups) > 0 && i.Flags&net.FlagMulticast != 0 {
				for _, multicastGroup := range multicastGroups {
					for _, port := range targetPorts {
						mapping[sourceAddr] = append(
							mapping[sourceAddr],
							&net.UDPAddr{
								IP:   multicastGroup,
								Port: port,
							},
						)
					}
				}
			}
		}
	}
	return
}
