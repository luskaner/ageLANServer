package models

import (
	"fmt"
	"io"
	"iter"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/server/internal"
)

func localIp(r *http.Request) (ip string) {
	addr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return
	}
	var err error
	ip, _, err = net.SplitHostPort(addr.String())
	if err != nil {
		return
	}
	if parsedIP := net.ParseIP(ip); parsedIP != nil && parsedIP.To4() != nil {
		return ip
	}
	return
}

var localSubnets []*net.IPNet
var publicIp string

func CacheNetworkInterfaces() {
	if internal.Connectivity {
		if resp, err := http.Get("https://api.ipify.org/"); err == nil {
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			if ipBytes, err := io.ReadAll(resp.Body); err == nil {
				ipStr := string(ipBytes)
				if ip := net.ParseIP(ipStr); ip != nil && ip.To4() != nil {
					publicIp = ipStr
				}
			}
		}
		if publicIp != "" {
			if ifs, err := common.RunningNetworkInterfaces(); err == nil {
				for _, ipNets := range ifs {
					for _, ipNet := range ipNets {
						localSubnets = append(localSubnets, ipNet)
					}
				}
			}
		}
	}
}

type BattleServer interface {
	SetLAN(lan bool)
	SetIPv4(ipv4 string)
	SetBsPort(bsPort int)
	SetWebSocketPort(webSocketPort int)
	SetOutOfBandPort(outOfBandPort int)
	SetHasOobPort(hasOobPort bool)
	SetBattleServerName(battleServerName string)
	SetName(name string)
	LAN() bool
	Region() string
	AppendName(encoded *internal.A)
	EncodeLogin(r *http.Request) internal.A
	EncodePorts() internal.A
	EncodeAdvertisement(r *http.Request) internal.A
	ResolveIPv4(r *http.Request) string
	String() string
}

type MainBattleServer struct {
	battleServer.Base `koanf:",squash"`
	lan               *bool
	hasOobPort        bool
	battleServerName  string
	lanMu             sync.RWMutex
}

func (battleServer *MainBattleServer) SetBattleServerName(battleServerName string) {
	battleServer.battleServerName = battleServerName
}

func (battleServer *MainBattleServer) SetHasOobPort(hasOobPort bool) {
	battleServer.hasOobPort = hasOobPort
}

func (battleServer *MainBattleServer) SetIPv4(ipv4 string) {
	battleServer.IPv4 = ipv4
}

func (battleServer *MainBattleServer) SetBsPort(bsPort int) {
	battleServer.BsPort = bsPort
}

func (battleServer *MainBattleServer) SetWebSocketPort(webSocketPort int) {
	battleServer.WebSocketPort = webSocketPort
}

func (battleServer *MainBattleServer) SetOutOfBandPort(outOfBandPort int) {
	battleServer.OutOfBandPort = outOfBandPort
}

func (battleServer *MainBattleServer) SetName(name string) {
	battleServer.Name = name
}

func (battleServer *MainBattleServer) LAN() bool {
	battleServer.lanMu.RLock()
	if battleServer.lan == nil {
		battleServer.lanMu.RUnlock()
		var lan bool
		battleServer.lanMu.Lock()
		battleServer.lan = &lan
		defer battleServer.lanMu.Unlock()
		if guid, err := uuid.Parse(battleServer.Base.Region); err == nil && guid.Version() == 4 {
			lan = true
		}
	} else {
		defer battleServer.lanMu.RUnlock()
	}
	return *battleServer.lan
}

func (battleServer *MainBattleServer) AppendName(encoded *internal.A) {
	switch battleServer.battleServerName {
	case "omit":
	case "null":
		*encoded = append(*encoded, nil)
	default:
		*encoded = append(*encoded, battleServer.Name)
	}
}

func (battleServer *MainBattleServer) SetLAN(enable bool) {
	battleServer.lanMu.Lock()
	defer battleServer.lanMu.Unlock()
	battleServer.lan = &enable
}

func (battleServer *MainBattleServer) Region() string {
	return battleServer.Base.Region
}

func (battleServer *MainBattleServer) EncodeLogin(r *http.Request) internal.A {
	encoded := internal.A{
		battleServer.Base.Region,
	}
	battleServer.AppendName(&encoded)
	encoded = append(encoded, battleServer.ResolveIPv4(r))
	encoded = append(encoded, battleServer.EncodePorts()...)
	return encoded
}

func (battleServer *MainBattleServer) EncodePorts() internal.A {
	encoded := internal.A{battleServer.BsPort}
	encoded = append(encoded, battleServer.WebSocketPort)
	if battleServer.hasOobPort {
		encoded = append(encoded, battleServer.OutOfBandPort)
	}
	return encoded
}

func (battleServer *MainBattleServer) EncodeAdvertisement(r *http.Request) internal.A {
	encoded := internal.A{
		battleServer.ResolveIPv4(r),
	}
	encoded = append(encoded, battleServer.EncodePorts()...)
	return encoded
}

func (battleServer *MainBattleServer) ResolveIPv4(r *http.Request) (ipV4 string) {
	if battleServer.IPv4 != "auto" {
		return battleServer.IPv4
	}
	remoteIPStr, _, _ := net.SplitHostPort(r.RemoteAddr)
	remoteIP := net.ParseIP(remoteIPStr)
	if remoteIP == nil || remoteIP.To4() == nil {
		return
	}
	var useLocalIp bool
	if internal.Connectivity {
		for _, subnet := range localSubnets {
			if subnet.Contains(remoteIP) {
				useLocalIp = true
				break
			}
		}
	} else {
		useLocalIp = true
	}
	if useLocalIp {
		return localIp(r)
	}
	host := r.Host
	if strings.Contains(host, ":") {
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}
	if ip := net.ParseIP(host); ip != nil && ip.To4() != nil {
		ipV4 = host
	} else {
		ipV4 = publicIp
	}
	return
}

func (battleServer *MainBattleServer) String() string {
	str := fmt.Sprintf(
		"Region: %s (Name: %s), IPv4: %s, Ports: ",
		battleServer.Base.Region,
		battleServer.Name,
		battleServer.IPv4,
	)
	ports := battleServer.EncodePorts()
	str += fmt.Sprintf("%v", ports)
	return str
}

type BattleServers interface {
	Initialize(battleServers []BattleServer, opts *BattleServerOpts)
	Iter() iter.Seq2[string, BattleServer]
	Encode(r *http.Request) internal.A
	Get(region string) (BattleServer, bool)
	NewLANBattleServer(region string) BattleServer
	NewBattleServer(region string) BattleServer
}

type BattleServerOpts struct {
	OobPort bool
	Name    string
}

type MainBattleServers struct {
	store            *internal.ReadOnlyOrderedMap[string, BattleServer]
	haveOobPort      bool
	battleServerName string
}

func (battleSrvs *MainBattleServers) Initialize(battleServers []BattleServer, opts *BattleServerOpts) {
	if opts == nil {
		opts = &BattleServerOpts{
			OobPort: true,
		}
	}
	if opts.Name == "" {
		opts.Name = "true"
	}
	keyOrder := make([]string, len(battleServers))
	mapping := make(map[string]BattleServer, len(battleServers))
	for i, bs := range battleServers {
		battleServers[i].SetHasOobPort(opts.OobPort)
		battleServers[i].SetBattleServerName(opts.Name)
		keyOrder[i] = bs.Region()
		mapping[keyOrder[i]] = battleServers[i]
	}
	battleSrvs.battleServerName = opts.Name
	battleSrvs.haveOobPort = opts.OobPort
	battleSrvs.store = internal.NewReadOnlyOrderedMap[string, BattleServer](keyOrder, mapping)
}

func (battleSrvs *MainBattleServers) Iter() iter.Seq2[string, BattleServer] {
	return battleSrvs.store.Iter()
}

func (battleSrvs *MainBattleServers) Encode(r *http.Request) internal.A {
	encoded := make(internal.A, battleSrvs.store.Len())
	i := 0
	for _, bs := range battleSrvs.store.Iter() {
		encoded[i] = bs.EncodeLogin(r)
		i++
	}
	return encoded
}

func (battleSrvs *MainBattleServers) Get(region string) (BattleServer, bool) {
	return battleSrvs.store.Load(region)
}

func (battleSrvs *MainBattleServers) NewLANBattleServer(region string) BattleServer {
	bs := battleSrvs.NewBattleServer(region)
	bs.SetLAN(true)
	return bs
}

func (battleSrvs *MainBattleServers) NewBattleServer(region string) BattleServer {
	return &MainBattleServer{
		Base: battleServer.Base{
			Region: region,
		},
		hasOobPort:       battleSrvs.haveOobPort,
		battleServerName: battleSrvs.battleServerName,
	}
}
