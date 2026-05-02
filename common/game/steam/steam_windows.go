package steam

import (
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func read(k registry.Key, path string, val string, _64bitView *bool) (value string) {
	var access uint32 = registry.QUERY_VALUE
	if _64bitView != nil {
		if *_64bitView {
			access |= registry.WOW64_64KEY
		} else {
			access |= registry.WOW64_32KEY
		}
	}
	key, err := registry.OpenKey(k, path, access)
	if err != nil {
		return
	}
	defer func(key registry.Key) {
		_ = key.Close()
	}(key)
	value, _, err = key.GetStringValue(val)
	if err != nil {
		value = ""
	}
	return
}

func ConfigPath() (path string) {
	return read(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, `SteamPath`, nil)
}

func ConfigPathAlt() (path string) {
	if p := read(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Steam`, `UninstallString`, new(false)); p != "" {
		path = filepath.Dir(p)
	}
	return
}
