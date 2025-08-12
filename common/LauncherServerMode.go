package common

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
)

type LauncherServerMode string

const (
	ModeQueryOrRun LauncherServerMode = ""
	ModeConnect    LauncherServerMode = "connect"
	ModeRun        LauncherServerMode = "run"
)

var supportedLauncherServerModes = mapset.NewThreadUnsafeSet(ModeQueryOrRun, ModeConnect, ModeRun)

func (s *LauncherServerMode) Set(val string) error {
	mode := LauncherServerMode(val)
	if !supportedLauncherServerModes.Contains(mode) {
		return fmt.Errorf("invalid launcher server mode: %s", mode)
	}
	*s = mode
	return nil
}

func (s *LauncherServerMode) Type() string {
	return "LauncherServerMode"
}

func (s *LauncherServerMode) String() string {
	return string(*s)
}

func (s *LauncherServerMode) UnmarshalText(text []byte) error {
	return s.Set(string(text))
}
