package game

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	internalExecutor "launcher/internal/executor"
	"log"
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
	// Does not seem there is another way without cgo?
	return executor.RunCustomExecutable("powershell", "-Command", "if ((Get-AppxPackage).Name -eq 'Microsoft.MSPhoenix') { exit 0 } else { exit 1 }")
}

func RunOnMicrosoftStore() bool {
	return internalExecutor.ShellExecute("open", `shell:appsfolder\Microsoft.MSPhoenix_8wekyb3d8bbwe!App`, false, windows.SW_HIDE) == nil
}

func RunOnSteam() bool {
	return internalExecutor.ShellExecute("open", "steam://rungameid/"+steamAppID, false, windows.SW_HIDE) == nil
}

func RunGame(executable string, userOnlyCertificate bool) bool {
	if executable != "auto" {
		switch executable {
		case "steam":
			if isInstalledOnSteam() {
				log.Println("AoE2:DE installed on Steam, launching...")
				return RunOnSteam()
			}
			return false
		case "msstore":
			if isInstalledOnMicrosoftStore() {
				log.Println("AoE2:DE installed on Microsoft Store, launching...")
				return RunOnMicrosoftStore()
			}
			return false
		default:
			log.Println(fmt.Sprintf("AoE2:DE launching custom launcher «%s»", executable))
			err, _ := internalExecutor.StartCustomExecutable(executable, true)
			if errors.Is(err, windows.ERROR_ELEVATION_REQUIRED) {
				log.Println("AoE2:DE Elevation required, retrying with admin privileges, accept any dialog if it appears...")
				if userOnlyCertificate {
					log.Println("Using a local user certificate. If it fails to connect to the server, try setting the config setting 'CanTrustCertificate' to 'local'.")
				}
				err = internalExecutor.ShellExecute("runas", executable, true, windows.SW_NORMAL)
			}
			if err != nil {
				log.Println("Failed to start custom launcher: " + err.Error())
				return false
			}
			return true
		}
	}
	if isInstalledOnSteam() {
		log.Println("AoE2:DE installed on Steam, launching...")
		return RunOnSteam()
	}
	if isInstalledOnMicrosoftStore() {
		log.Println("AoE2:DE installed on Microsoft Store, launching...")
		return RunOnMicrosoftStore()
	}
	return false
}
