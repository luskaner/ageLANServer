package server

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"golang.org/x/net/ipv4"
)

const latencyMeasurementCount = 3

type MesuredIpAddress struct {
	Ip      net.IP
	Latency time.Duration
}

func StartServer(gameTitle string, stop string, executable string, args []string, id uuid.UUID, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result, executablePath string, ip string) {
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
	options := commonExecutor.Options{File: executablePath, Args: args, ShowWindow: showWindow, Pid: true}
	optionsFn(options)
	result = options.Exec()
	if result.Success() {
		localIPs := common.NetIPSliceToNetIPSet(common.StringSliceToNetIPSlice(common.HostOrIpToIps(netip.IPv4Unspecified().String())))
		timeout := time.After(time.Duration(localIPs.Cardinality()) * (latencyMeasurementCount + 1) * time.Second)
	loop:
		for {
			select {
			case <-timeout:
				break loop
			default:
				if _, measuredIpAddrs, data := FilterServerIPs(id, "", gameTitle, localIPs); data != nil {
					ip = measuredIpAddrs[0].Ip.String()
					return
				}
				if _, err := commonProcess.FindProcess(int(result.Pid)); err != nil {
					break loop
				}
			}
		}
		if _, proc, err := commonProcess.Process(executablePath); err == nil && proc != nil {
			if err = serverKill.Do(executablePath); err != nil {
				logger.Println("Failed to stop 'server'")
				logger.Println("Error message: " + err.Error())
				logger.Println("You may try killing it manually. Kill process 'server' in your task manager.")
			}
		}
		result = nil
	}
	return
}

func GenerateServerCertificates(serverExecutablePath string, canTrustCertificate bool) (errorCode int) {
	if exists, certificateFolder, cert, _, caCert, selfSignedCert, _ := common.CertificatePairs(serverExecutablePath); !exists || CertificateSoonExpired(cert) || CertificateSoonExpired(caCert) || CertificateSoonExpired(selfSignedCert) {
		if !canTrustCertificate {
			logger.Println("serverStart is true and canTrustCertificate is false. Certificate pair is missing or soon expired. Generate your own certificates manually.")
			errorCode = internal.ErrServerCertMissingExpired
			return
		}
		if certificateFolder == "" {
			logger.Println("Cannot find certificate folder of the 'server'. Make sure the folder structure of the 'server' is correct.")
			errorCode = internal.ErrServerCertDirectory
			return
		}
		if result := GenerateCertificatePair(certificateFolder, func(options commonExecutor.Options) {

		}); !result.Success() {
			logger.Println("Failed to generate certificate pair. Check the folder and its permissions")
			errorCode = internal.ErrServerCertCreate
			if result.Err != nil {
				logger.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			return
		}
	}
	return
}

func GetExecutablePath(executable string) string {
	if executable == "auto" {
		return executables.FindPath(executables.Filename(true, executables.Server))
	}
	return executable
}

func LanServerHost(id uuid.UUID, gameTitle string, host string, insecureSkipVerify bool) (ok bool) {
	ipAddrs := common.HostToIps(host)
	if len(ipAddrs) == 0 {
		return
	}
	for _, ipAddr := range ipAddrs {
		if ok, _, _, _ = lanServerIP(id, gameTitle, ipAddr, host, insecureSkipVerify, true); !ok {
			return
		}
	}
	return true
}

func FilterServerIPs(id uuid.UUID, serverName string, gameTitle string, possibleIpAddrs mapset.Set[netip.Addr]) (actualId uuid.UUID, measuredIpAddresses []MesuredIpAddress, data *AnnounceMessageDataSupportedLatest) {
	for ipAddr := range possibleIpAddrs.Iter() {
		ip := common.NetIPAddrToNetIP(ipAddr)
		var ok bool
		var latency time.Duration
		var tmpData *AnnounceMessageDataSupportedLatest
		if ok, actualId, latency, tmpData = lanServerIP(id, gameTitle, ip, serverName, true, false); ok {
			measuredIpAddresses = append(measuredIpAddresses, MesuredIpAddress{
				Ip:      ip,
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

func lanServerIP(id uuid.UUID, gameTitle string, ipAddr net.IP, serverName string, insecureSkipVerify bool, ignoreLatency bool) (ok bool, serverId uuid.UUID, latency time.Duration, data *AnnounceMessageDataSupportedLatest) {
	tr := &http.Transport{
		TLSClientConfig: TlsConfig(serverName, insecureSkipVerify),
	}
	client := &http.Client{Transport: tr, Timeout: 1 * time.Second}
	u := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(ipAddr.String(), ""),
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
	serverIdStr := resp.Header.Get(common.IdHeader)
	if version == "" || serverIdStr == "" {
		return
	}
	versionInt, _ := strconv.Atoi(version)
	if versionInt > common.AnnounceVersionLatest {
		return
	}
	var serverIdUuid uuid.UUID
	serverIdUuid, err = uuid.Parse(serverIdStr)
	if err != nil {
		return
	}
	if id != uuid.Nil && id != serverIdUuid {
		return
	}
	serverId = serverIdUuid
	data = &AnnounceMessageDataSupportedLatest{}
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return
	}
	if data.GameTitle != gameTitle {
		return
	}
	ok = true
	return
}

func QueryServers(
	multicastGroups mapset.Set[netip.Addr],
	targetPorts mapset.Set[uint16],
	servers map[uuid.UUID]*AnnounceMessage,
) {
	sourceToTargetAddrs := sourceToTargetUDPAddrs(
		multicastGroups,
		targetPorts,
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
					IpAddrs: mapset.NewThreadUnsafeSet[netip.Addr](),
				}
				servers[parsedId] = server
			}
			server.IpAddrs.Add(common.NetIPToNetIPAddr(addr.IP))
		}()
	}

	var wg sync.WaitGroup
	for _, conn := range connTargets {
		wg.Add(1)
		go func(conn *connTarget) {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			packetBuffer := make([]byte, len(common.AnnounceHeader)+AnnounceIdLength)
			for i := 0; i < 3; i++ {
				select {
				case <-ticker.C:
					sendAndReceive(&packetBuffer, conn, servers)
				}
			}
			wg.Done()
		}(conn)
	}
	wg.Wait()
}

func calculateBroadcastIPv4(ip net.IP, mask net.IPMask) net.IP {
	broadcast := make(net.IP, len(ip))
	for i, b := range ip {
		broadcast[i] = b | ^mask[i]
	}
	return broadcast
}

func sourceToTargetUDPAddrs(
	multicastGroups mapset.Set[netip.Addr],
	targetPorts mapset.Set[uint16],
) (mapping map[*net.UDPAddr][]*net.UDPAddr) {
	interfaces, err := common.RunningNetworkInterfaces()
	if err != nil {
		return nil
	}
	mapping = make(map[*net.UDPAddr][]*net.UDPAddr)
	for iff, iffIps := range interfaces {
		for _, n := range iffIps {
			sourceAddr := &net.UDPAddr{
				IP: n.IP,
			}
			mapping[sourceAddr] = make([]*net.UDPAddr, 0)
			if iff.Flags&net.FlagBroadcast != 0 {
				for port := range targetPorts.Iter() {
					mapping[sourceAddr] = append(mapping[sourceAddr], &net.UDPAddr{
						IP:   calculateBroadcastIPv4(sourceAddr.IP.To4(), n.Mask),
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
								IP:   multicastGroup.AsSlice(),
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
