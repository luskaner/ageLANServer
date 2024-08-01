package game

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	commonExecutor "launcherCommon/executor"
	"os"
)

type Executor interface {
	Execute(args []string) (result *commonExecutor.ExecResult)
	GameProcesses() (steamProcess bool, microsoftStoreProcess bool)
}

type SteamExecutor struct{}
type MicrosoftStoreExecutor struct{}
type CustomExecutor struct {
	Executable string
}

const steamAppID = "813780"

func (exec SteamExecutor) Execute(_ []string) (result *commonExecutor.ExecResult) {
	result = commonExecutor.ExecOptions{File: "steam://rungameid/" + steamAppID, Shell: true, SpecialFile: true}.Exec()
	return
}

func (exec SteamExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	return
}

func (exec MicrosoftStoreExecutor) Execute(_ []string) (result *commonExecutor.ExecResult) {
	result = commonExecutor.ExecOptions{File: `shell:appsfolder\Microsoft.MSPhoenix_8wekyb3d8bbwe!App`, Shell: true, SpecialFile: true}.Exec()
	return
}

func (exec MicrosoftStoreExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	microsoftStoreProcess = true
	return
}

func (exec CustomExecutor) Execute(args []string) (result *commonExecutor.ExecResult) {
	result = commonExecutor.ExecOptions{File: exec.Executable, Args: args}.Exec()
	return
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	microsoftStoreProcess = true
	return
}

func (exec CustomExecutor) ExecuteElevated(args []string) (result *commonExecutor.ExecResult) {
	result = commonExecutor.ExecOptions{File: exec.Executable, AsAdmin: true, WindowState: windows.SW_NORMAL, Args: args}.Exec()
	return
}

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
	return commonExecutor.ExecOptions{File: "powershell", SpecialFile: true, Wait: true, ExitCode: true, Args: []string{"-Command", "if ((Get-AppxPackage).Name -eq 'Microsoft.MSPhoenix') { exit 0 } else { exit 1 }"}}.Exec().Success()
}

func isInstalledCustom(executable string) bool {
	info, err := os.Stat(executable)
	if err != nil || os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

func MakeExecutor(executable string) Executor {
	if executable != "auto" {
		switch executable {
		case "steam":
			if isInstalledOnSteam() {
				return SteamExecutor{}
			}
		case "msstore":
			if isInstalledOnMicrosoftStore() {
				return MicrosoftStoreExecutor{}
			}
		default:
			if isInstalledCustom(executable) {
				return CustomExecutor{Executable: executable}
			}
		}
		return nil
	}
	if isInstalledOnSteam() {
		return SteamExecutor{}
	}
	if isInstalledOnMicrosoftStore() {
		return MicrosoftStoreExecutor{}
	}
	return nil
}
