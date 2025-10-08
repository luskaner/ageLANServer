package models

import (
	"fmt"
	"iter"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/server/internal"
)

type MainBattleServer struct {
	battleServerConfig.BaseConfig `mapstructure:",squash"`
	lan                           *bool
	hasOobPort                    bool
	battleServerName              string
}

func (battleServer *MainBattleServer) LAN() bool {
	if battleServer.lan == nil {
		var lan bool
		battleServer.lan = &lan
		if guid, err := uuid.Parse(battleServer.Region); err == nil && guid.Version() == 4 {
			lan = true
		}
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

func (battleServer *MainBattleServer) EncodeLogin(r *http.Request) internal.A {
	encoded := internal.A{
		battleServer.Region,
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
		battleServer.Region,
		battleServer.Name,
		battleServer.IPv4,
	)
	ports := battleServer.EncodePorts()
	str += fmt.Sprintf("%v", ports)
	return str
}

type MainBattleServers struct {
	store            *internal.ReadOnlyOrderedMap[string, *MainBattleServer]
	haveOobPort      bool
	battleServerName string
}

func (battleSrvs *MainBattleServers) Initialize(battleServers []MainBattleServer, haveOobPort bool, battleServerName string) {
	keyOrder := make([]string, len(battleServers))
	mapping := make(map[string]*MainBattleServer, len(battleServers))
	for i, bs := range battleServers {
		battleServers[i].hasOobPort = haveOobPort
		battleServers[i].battleServerName = battleServerName
		keyOrder[i] = bs.Region
		mapping[keyOrder[i]] = &battleServers[i]
	}
	battleSrvs.battleServerName = battleServerName
	battleSrvs.haveOobPort = haveOobPort
	battleSrvs.store = internal.NewReadOnlyOrderedMap[string, *MainBattleServer](keyOrder, mapping)
}

func (battleSrvs *MainBattleServers) Iter() iter.Seq2[string, *MainBattleServer] {
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

func (battleSrvs *MainBattleServers) Get(region string) (*MainBattleServer, bool) {
	return battleSrvs.store.Load(region)
}

func (battleSrvs *MainBattleServers) NewLANBattleServer(region string) *MainBattleServer {
	battleServer := battleSrvs.NewBattleServer(region)
	lan := true
	battleServer.lan = &lan
	return battleServer
}

func (battleSrvs *MainBattleServers) NewBattleServer(region string) *MainBattleServer {
	return &MainBattleServer{
		BaseConfig: battleServerConfig.BaseConfig{
			Region: region,
		},
		hasOobPort:       battleSrvs.haveOobPort,
		battleServerName: battleSrvs.battleServerName,
	}
}
