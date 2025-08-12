package common

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"runtime"
)

type LauncherUserStore string

const (
	StoreUser LauncherUserStore = "user"
)

var supportedLauncherUserStores mapset.Set[LauncherUserStore]

func init() {
	supportedLauncherUserStores = mapset.NewThreadUnsafeSet[LauncherUserStore](
		LauncherUserStore(StoreNone),
		LauncherUserStore(StoreLocal),
		LauncherUserStore(StoreTmp),
	)
	if runtime.GOOS == "windows" {
		supportedLauncherUserStores.Add(StoreUser)
	}
}

func (s *LauncherUserStore) IsTmp() bool {
	return *s == LauncherUserStore(StoreTmp)
}

func (s *LauncherUserStore) IsNone() bool {
	return *s == LauncherUserStore(StoreNone)
}

func (s *LauncherUserStore) Set(val string) error {
	store := LauncherUserStore(val)
	if !supportedLauncherUserStores.Contains(store) {
		return fmt.Errorf("invalid launcher store: %s", store)
	}
	*s = store
	return nil
}

func (s *LauncherUserStore) Type() string {
	return "LauncherUserStore"
}

func (s *LauncherUserStore) String() string {
	return string(*s)
}

func (s *LauncherUserStore) UnmarshalText(text []byte) error {
	return s.Set(string(text))
}
