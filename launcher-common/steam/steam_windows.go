package steam

import "golang.org/x/sys/windows/registry"

func HomeDirPath() (path string) {
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
