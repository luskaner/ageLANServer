package game

import (
	"golang.org/x/sys/windows/registry"
	"shared/executor"
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
	return executor.RunCustomExecutable("powershell", "-Command", "if ((Get-AppxPackage).Name -eq 'Microsoft.MSPhoenix') { exit 0 } else { exit 1 }")
}

func RunOnMicrosoftStore() bool {
	return executor.StartCustomExecutable(`explorer.exe`, false, `shell:appsfolder\Microsoft.MSPhoenix_8wekyb3d8bbwe!App`) != nil
}

func RunOnSteam() bool {
	return executor.StartCustomExecutable("rundll32.exe", false, "url.dll,FileProtocolHandler", "steam://rungameid/"+steamAppID) != nil
}

func RunGame(executable string) bool {
	if executable != "auto" {
		switch executable {
		case "steam":
			if isInstalledOnSteam() {
				return RunOnSteam()
			}
			return false
		case "msstore":
			if isInstalledOnMicrosoftStore() {
				return RunOnMicrosoftStore()
			}
			return false
		default:
			return executor.StartCustomExecutable(executable, true) != nil
		}
	}
	if isInstalledOnSteam() {
		return RunOnSteam()
	}
	if isInstalledOnMicrosoftStore() {
		return RunOnMicrosoftStore()
	}
	return false
}