package common

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
)

type LauncherStore string

const (
	StoreNone  LauncherStore = ""
	StoreLocal LauncherStore = "local"
	StoreTmp   LauncherStore = "tmp"
)

var supportedLauncherStores = mapset.NewThreadUnsafeSet(StoreNone, StoreLocal, StoreTmp)

func (s *LauncherStore) Set(val string) error {
	store := LauncherStore(val)
	if !supportedLauncherStores.Contains(store) {
		return fmt.Errorf("invalid launcher store: %s", store)
	}
	*s = store
	return nil
}

func (s *LauncherStore) Type() string {
	return "LauncherStore"
}

func (s *LauncherStore) String() string {
	return string(*s)
}

func (s *LauncherStore) UnmarshalText(text []byte) error {
	return s.Set(string(text))
}
