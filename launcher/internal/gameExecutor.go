package internal

import (
	"golang.org/x/sys/windows/registry"
)

const steamAppID = "813780"

func isInstalledOnSteam() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam\Apps\`+steamAppID, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer func(key registry.Key) {
		_ = key.Close()
	}(key)
	val, _, err := key.GetIntegerValue("Installed")
	if err != nil {
		return false
	}
	return val == 1
}

func isInstalledOnMicrosoftStore() bool {
	return RunCustomExecutable("powershell", "-Command", "if ((Get-AppxPackage).Name -eq 'Microsoft.MSPhoenix') { exit 0 } else { exit 1 }")
}

func RunOnMicrosoftStore() bool {
	return StartCustomExecutable(`explorer.exe`, `shell:appsfolder\Microsoft.MSPhoenix_8wekyb3d8bbwe!App`) != nil
}

func RunOnSteam() bool {
	return StartCustomExecutable("rundll32.exe", "url.dll,FileProtocolHandler", "steam://rungameid/"+steamAppID) != nil
}

func RunGame(config ClientConfig) bool {
	if config.Executable != "" {
		return StartCustomExecutable(config.Executable) != nil
	}
	if isInstalledOnSteam() {
		return RunOnSteam()
	}
	if isInstalledOnMicrosoftStore() {
		return RunOnMicrosoftStore()
	}
	return false
}
