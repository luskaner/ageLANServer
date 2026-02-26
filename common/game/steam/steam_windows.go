package steam

import (
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func ConfigPath() (path string) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer func(key registry.Key) {
		_ = key.Close()
	}(key)
	var val string
	val, _, err = key.GetStringValue("SteamPath")
	if err != nil {
		return
	}
	return val
}

func ConfigPathAlt() (path string) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer func(key registry.Key) {
		_ = key.Close()
	}(key)
	var val string
	val, _, err = key.GetStringValue("UninstallString")
	if err != nil {
		return
	}
	return filepath.Dir(val)
}
