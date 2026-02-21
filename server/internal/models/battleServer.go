package models

import (
	"fmt"
	"iter"
	"net"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/server/internal"
)

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
	battleServerConfig.BaseConfig `mapstructure:",squash"`
	lan                           *bool
	hasOobPort                    bool
	battleServerName              string
	lanMu                         sync.RWMutex
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
		if guid, err := uuid.Parse(battleServer.BaseConfig.Region); err == nil && guid.Version() == 4 {
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
	return battleServer.BaseConfig.Region
}

func (battleServer *MainBattleServer) EncodeLogin(r *http.Request) internal.A {
	encoded := internal.A{
		battleServer.BaseConfig.Region,
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

func (battleServer *MainBattleServer) ResolveIPv4(r *http.Request) string {
	if battleServer.IPv4 == "auto" {
		addr, _ := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
		ip, _, _ := net.SplitHostPort(addr.String())
		return ip
	}
	return battleServer.IPv4
}

func (battleServer *MainBattleServer) String() string {
	str := fmt.Sprintf(
		"Region: %s (Name: %s), IPv4: %s, Ports: ",
		battleServer.BaseConfig.Region,
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
	battleServer := battleSrvs.NewBattleServer(region)
	battleServer.SetLAN(true)
	return battleServer
}

func (battleSrvs *MainBattleServers) NewBattleServer(region string) BattleServer {
	return &MainBattleServer{
		BaseConfig: battleServerConfig.BaseConfig{
			Region: region,
		},
		hasOobPort:       battleSrvs.haveOobPort,
		battleServerName: battleSrvs.battleServerName,
	}
}
