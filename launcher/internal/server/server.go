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
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
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

const latencyMeasurementCount = 3

type MesuredIpAddress struct {
	IpAddr  netip.Addr
	Latency time.Duration
}

func StartServer(gameTitle common.GameTitle, id uuid.UUID, stop string, executablePath string, args []string, ipProtocol *common.IPProtocol) (result *commonExecutor.Result, ipAddr netip.Addr) {
	var host string
	var ipProtocolServer common.IPProtocol
	if ipProtocol.IPv4() {
		host = netip.IPv4Unspecified().String()
		ipProtocolServer = common.IPv4
	} else {
		host = netip.IPv6Unspecified().String()
		ipProtocolServer = common.IPv6
	}
	env := executor.EnvMap(
		common.Server,
		map[string]string{
			"GameTitles":                  string(gameTitle),
			"Network.Listen.DisableHttps": "0",
			"Network.Listen.Port":         "https",
			"Network.Listen.Hosts.Values": host,
			"Network.IPProtocol":          string(ipProtocolServer),
			"Id":                          id.String(),
		},
	)
	result = commonExecutor.Options{
		File:       executablePath,
		Args:       args,
		ShowWindow: stop != "true",
		Pid:        true,
		Env:        env,
	}.Exec()
	if result.Success() {
		localIPs := launcherCommon.AddrToIpAddrs(host, ipProtocol.IPv4(), !ipProtocol.IPv4())
		timeout := time.After(time.Duration(localIPs.Cardinality()) * (latencyMeasurementCount + 1) * time.Second)
	loop:
		for {
			select {
			case <-timeout:
				break loop
			default:
				if measuredIpAddrs, data := FilterServerIPs(id, "", gameTitle, localIPs); data != nil {
					ipAddr = measuredIpAddrs[0].IpAddr
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

func LanServerHost(id uuid.UUID, gameTitle common.GameTitle, host string, insecureSkipVerify bool, IPProtocol *common.IPProtocol) (ok bool) {
	ipAddrs := launcherCommon.AddrToIpAddrs(host, IPProtocol.IPv4(), IPProtocol.IPv6())
	if ipAddrs.IsEmpty() {
		return
	}
	for ipAddr := range ipAddrs.Iter() {
		if ok, _, _ = LanServerIP(id, gameTitle, ipAddr, host, insecureSkipVerify, true); !ok {
			return
		}
	}
	return true
}

func LanServerIP(id uuid.UUID, gameTitle common.GameTitle, ipAddr netip.Addr, serverName string, insecureSkipVerify bool, ignoreLatency bool) (ok bool, latency time.Duration, data *AnnounceMessageDataSupportedLatest) {
	tr := &http.Transport{
		TLSClientConfig: TlsConfig(serverName, insecureSkipVerify),
	}
	client := &http.Client{Transport: tr, Timeout: 1 * time.Second}
	u := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(ipAddr.String(), "https"),
		Path:   "test",
	}
	if !ignoreLatency {
		var latencyAccumulator time.Duration
		for i := 0; i < latencyMeasurementCount; i++ {
			start := time.Now()
			req, err := http.NewRequest("HEAD", u.String(), nil)
			if err != nil {
				return
			}
			req.Host = serverName
			if _, err = client.Do(req); err != nil {
				return
			}
			latencyAccumulator += time.Since(start)
		}
		latency = latencyAccumulator / latencyMeasurementCount
	}
	req, err := http.NewRequest(
		"GET",
		u.String(),
		nil,
	)
	if err != nil {
		return
	}
	req.Host = serverName
	resp, err := client.Do(req)
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
	if id != uuid.Nil && id != serverIdUuid {
		return
	}
	data = &AnnounceMessageDataSupportedLatest{}
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return
	}
	if !slices.Contains(data.GameTitles, string(gameTitle)) {
		return
	}
	ok = true
	return
}

func FilterServerIPs(id uuid.UUID, serverName string, gameTitle common.GameTitle, possibleIpAddrs mapset.Set[netip.Addr]) (measuredIpAddresses []MesuredIpAddress, data *AnnounceMessageDataSupportedLatest) {
	for ipAddr := range possibleIpAddrs.Iter() {
		if ok, latency, tmpData := LanServerIP(id, gameTitle, ipAddr, serverName, true, false); ok {
			measuredIpAddresses = append(measuredIpAddresses, MesuredIpAddress{
				IpAddr:  ipAddr,
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

func QueryServers(
	ctx context.Context,
	multicastGroupsIPv4 mapset.Set[netip.Addr],
	targetPortsIPv4 mapset.Set[uint16],
	multicastGroupsIPv6 mapset.Set[netip.Addr],
	targetPortsIPv6 mapset.Set[uint16],
	broadcastIpv4 bool,
	servers map[uuid.UUID]*AnnounceMessage,
	ipProtocol *common.IPProtocol) {
	actualMulticastGroupsIPv4 := multicastGroupsIPv4.Clone()
	actualTargetPortsIPv4 := targetPortsIPv4.Clone()
	actualMulticastGroupsIPv6 := multicastGroupsIPv6.Clone()
	actualTargetPortsIPv6 := targetPortsIPv6.Clone()
	actualBroadcastIpv4 := broadcastIpv4
	if !ipProtocol.IPv6() {
		actualMulticastGroupsIPv6.Clear()
		actualTargetPortsIPv6.Clear()
	}
	if !ipProtocol.IPv4() {
		actualMulticastGroupsIPv4.Clear()
		actualTargetPortsIPv4.Clear()
		actualBroadcastIpv4 = false
	}
	sourceToTargetAddrs := sourceToTargetUDPAddrs(
		actualMulticastGroupsIPv4,
		actualTargetPortsIPv4,
		actualMulticastGroupsIPv6,
		actualTargetPortsIPv6,
		actualBroadcastIpv4,
	)
	if len(sourceToTargetAddrs) == 0 {
		return
	}
	type connTarget struct {
		conn   *net.UDPConn
		target *net.UDPAddr
	}
	var connTargets []*connTarget
	for source, targets := range sourceToTargetAddrs {
		network := "udp"
		IPv4 := source.IP.To4() != nil
		if *ipProtocol != common.IPvDual {
			if IPv4 {
				network += "4"
			} else {
				network += "6"
			}
		}
		conn, err := net.ListenUDP(
			network,
			source,
		)
		if err != nil {
			continue
		}
		if slices.ContainsFunc(targets, func(addr *net.UDPAddr) bool {
			return addr.IP.IsMulticast()
		}) {
			if IPv4 {
				p := ipv4.NewPacketConn(conn)
				if err = p.SetMulticastLoopback(true); err != nil {
					continue
				}
			} else {
				p := ipv6.NewPacketConn(conn)
				if err = p.SetMulticastLoopback(true); err != nil {
					continue
				}
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
					IpAddrs: mapset.NewThreadUnsafeSet[netip.Addr](),
				}
				servers[parsedId] = server
			}
			server.IpAddrs.Add(common.NetIPToNetIPAddr(addr.IP).WithZone(addr.Zone))
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

func calculateBroadcastIPv4(ip net.IP, mask net.IPMask) net.IP {
	broadcast := make(net.IP, len(ip))
	for i, b := range ip {
		broadcast[i] = b | ^mask[i]
	}
	return broadcast
}

func sourceToTargetUDPAddrs(
	multicastGroupsIPv4 mapset.Set[netip.Addr],
	targetPortsIPv4 mapset.Set[uint16],
	multicastGroupsIPv6 mapset.Set[netip.Addr],
	targetPortsIPv6 mapset.Set[uint16],
	broadcastIPv4 bool,
) (mapping map[*net.UDPAddr][]*net.UDPAddr) {
	interfaces, err := common.RunningNetworkInterfaces(
		!targetPortsIPv4.IsEmpty(),
		!targetPortsIPv6.IsEmpty(),
		false,
	)
	if err != nil {
		return nil
	}
	mapping = make(map[*net.UDPAddr][]*net.UDPAddr)
	for iff, iffIps := range interfaces {
		for _, n := range iffIps {
			sourceAddr := &net.UDPAddr{
				IP: n.IP,
			}
			isIPv4 := n.IP.To4() != nil
			var targetPorts mapset.Set[uint16]
			var multicastGroups mapset.Set[netip.Addr]
			if isIPv4 {
				sourceAddr.IP = n.IP.To4()
				targetPorts = targetPortsIPv4
				multicastGroups = multicastGroupsIPv4
			} else {
				sourceAddr.IP = n.IP.To16()
				targetPorts = targetPortsIPv6
				multicastGroups = multicastGroupsIPv6
			}
			mapping[sourceAddr] = make([]*net.UDPAddr, 0)
			if isIPv4 && broadcastIPv4 && iff.Flags&net.FlagBroadcast != 0 {
				for port := range targetPorts.Iter() {
					mapping[sourceAddr] = append(mapping[sourceAddr], &net.UDPAddr{
						IP:   calculateBroadcastIPv4(sourceAddr.IP, n.Mask),
						Port: int(port),
					})
				}
			} else {
				for port := range targetPorts.Iter() {
					mapping[sourceAddr] = append(mapping[sourceAddr], &net.UDPAddr{
						IP:   sourceAddr.IP,
						Port: int(port),
					})
				}
			}

			if !multicastGroups.IsEmpty() && iff.Flags&net.FlagMulticast != 0 {
				for multicastGroup := range multicastGroups.Iter() {
					for port := range targetPorts.Iter() {
						mapping[sourceAddr] = append(
							mapping[sourceAddr],
							&net.UDPAddr{
								IP:   common.NetIPAddrToNetIP(multicastGroup),
								Port: int(port),
							},
						)
					}
				}
			}
		}
	}
	return
}
