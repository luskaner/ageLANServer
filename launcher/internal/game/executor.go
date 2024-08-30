package game

import (
	"github.com/andygrunwald/vdf"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"os"
	"path"
)

type Executor interface {
	Execute(args []string) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, microsoftStoreProcess bool)
}

type SteamExecutor struct{}
type MicrosoftStoreExecutor struct{}
type CustomExecutor struct {
	Executable string
}

const steamAppID = "813780"

func (exec SteamExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	return startUri("steam://rungameid/" + steamAppID)
}

func (exec SteamExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	return
}

func (exec CustomExecutor) Execute(args []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: exec.Executable, Args: args}.Exec()
	return
}

func (exec CustomExecutor) ExecuteElevated(args []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: exec.Executable, AsAdmin: true, ShowWindow: true, Args: args}.Exec()
	return
}

func isInstalledCustom(executable string) bool {
	info, err := os.Stat(executable)
	if err != nil || os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

func isInstalledOnSteam() bool {
	p := steamInstallationPath()
	if p == "" {
		return false
	}
	f, err := os.Open(path.Join(p, "config", "libraryfolders.vdf"))
	if err != nil {
		return false
	}
	parser := vdf.NewParser(f)
	data, err := parser.Parse()
	if err != nil {
		return false
	}
	libraryFolders, ok := data["libraryfolders"].(map[string]interface{})
	if !ok {
		return false
	}

	for _, folder := range libraryFolders {
		var folderMap map[string]interface{}
		folderMap, ok = folder.(map[string]interface{})
		if !ok {
			continue
		}
		var apps map[string]interface{}
		apps, ok = folderMap["apps"].(map[string]interface{})
		if !ok {
			continue
		}
		if _, exists := apps[steamAppID]; exists {
			return true
		}
	}
	return false
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
