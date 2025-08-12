package common

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/executor"
	"runtime"
)

type ClientLauncher string

const (
	ClientLauncherSteamOrMSStore ClientLauncher = ""
	ClientLauncherSteam          ClientLauncher = "steam"
	ClientLauncherMSStore        ClientLauncher = "msstore"
	ClientLauncherPath           ClientLauncher = "path"
)

var supportedClientLaunchers mapset.Set[ClientLauncher]

func init() {
	var supportsSteamOrMSStore bool
	supportedClientLaunchers = mapset.NewThreadUnsafeSet[ClientLauncher](
		ClientLauncherPath,
	)
	if runtime.GOOS == "windows" {
		supportedClientLaunchers.Add(ClientLauncherMSStore)
		supportsSteamOrMSStore = true
	}
	if runtime.GOOS == "linux" && !executor.IsAdmin() {
		supportsSteamOrMSStore = true
	}
	if supportsSteamOrMSStore {
		supportedClientLaunchers.Add(ClientLauncherSteamOrMSStore)
		supportedClientLaunchers.Add(ClientLauncherSteam)
	}
}

func (c *ClientLauncher) Set(val string) error {
	clientLauncher := ClientLauncher(val)
	if !supportedClientLaunchers.Contains(clientLauncher) {
		return fmt.Errorf("%v is not a supported ClientLauncher", val)
	}
	*c = clientLauncher
	return nil
}

func (c *ClientLauncher) Type() string {
	return "ClientLauncher"
}

func (c *ClientLauncher) String() string {
	return string(*c)
}

func (c *ClientLauncher) UnmarshalText(text []byte) error {
	return c.Set(string(text))
}
