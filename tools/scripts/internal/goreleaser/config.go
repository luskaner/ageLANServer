package goreleaser

import (
	"fmt"
	"path/filepath"
	"strings"
)

const configSource = `%s/resources/config.game.toml`
const scriptSource = `%s/resources/{{.BaseOS}}/%s.{{.SrcScriptExt}}`
const gameScriptSource = `%s/resources/{{.BaseOS}}/start_{{.Game}}.{{.SrcScriptExt}}`

func Generate() {
	// Server Archive
	serverArchive := NewArchive("server", Targets3264)
	serverArchive.AddDocFiles("docs", nil, nil, "LICENSE", "server/README.md", "server/BattleServers.md")
	serverArchive.AddSrcDstFile("server/resources/responses", "resources/responses")
	serverArchive.AddSrcDstFile("server/resources/config", "resources/config")
	serverArchive.AddScriptFiles("", NewTemplate[FileData](fmt.Sprintf(gameScriptSource, `server`)), nil, nil, true)
	serverArchive.AddScriptFiles("bin", NewTemplate[FileData](fmt.Sprintf(scriptSource, `server-genCert`, `genCert`)), SourceIgnoreFn{
		OSWindows: func(path string) bool {
			return path == `server-genCert/resources/windows/genCert.bat`
		},
		OSMacOS: func(path string) bool {
			return path == `server-genCert/resources/unix/genCert.sh`
		},
	}, nil, false)
	server := NewBinary("./server", Targets3264)
	serverArchive.AddMainBinary(server)
	serverGenCert := NewBinary("./server-genCert", Targets3264)
	serverArchive.AddAuxiliarBinary(serverGenCert)
	// Battle Server Manager Archive
	battleServerManagerArchive := NewArchive("battle-server-manager", Targets64ExceptMacOS)
	battleServerManagerArchive.AddDocFiles("docs", nil, nil, "battle-server-manager/README.md")
	battleServerManagerArchive.AddScriptFiles("", NewTemplate[FileData](fmt.Sprintf(gameScriptSource, `battle-server-manager`)), nil, nil, true)
	battleServerManagerArchive.AddScriptFiles("", NewTemplate[FileData](fmt.Sprintf(scriptSource, `battle-server-manager`, `clean`)), nil, nil, false)
	battleServerManagerArchive.AddScriptFiles("", NewTemplate[FileData](fmt.Sprintf(scriptSource, `battle-server-manager`, `remove-all`)), nil, nil, false)
	battleServerManagerArchive.AddConfigFiles("", NewTemplate[FileData](fmt.Sprintf(configSource, `battle-server-manager`)), true)
	battleServerManager := NewBinary("./battle-server-manager", Targets64ExceptMacOS)
	battleServerManagerArchive.AddMainBinary(battleServerManager)
	// Launcher archive
	launcherArchive := NewArchive("launcher", Targets64ExceptMacOS)
	launcherArchive.AddSrcDstFile("launcher/resources/config.toml", "resources/config.toml")
	launcherArchive.AddScriptFiles("", NewTemplate[FileData](fmt.Sprintf(gameScriptSource, `launcher`)), nil, nil, true)
	launcherArchive.AddConfigFiles("", NewTemplate[FileData](fmt.Sprintf(configSource, `launcher`)), true)
	launcherArchive.AddDocFiles("docs", nil, nil, "launcher/README.md", "LICENSE")
	launcherArchive.AddDocFiles("docs", func(source string) Renders[FileData] {
		ext := filepath.Ext(source)
		root := strings.TrimSuffix(source, ext)
		return LiteralString[FileData](fmt.Sprintf("%s%s%s", root, "-config", ext))
	}, nil, "launcher-config/README.md")
	launcher := NewBinary("./launcher", Targets64ExceptMacOS)
	launcherArchive.AddMainBinary(launcher)
	launcherAgent := NewBinary("./launcher-agent", Targets64ExceptMacOS)
	launcherArchive.AddAuxiliarBinary(launcherAgent)
	launcherConfig := NewBinary("./launcher-config", Targets64ExceptMacOS)
	launcherArchive.AddAuxiliarBinary(launcherConfig)
	launcherConfigAdmin := NewBinary("./launcher-config-admin", Targets64ExceptMacOS)
	launcherArchive.AddAuxiliarBinary(launcherConfigAdmin)
	launcherConfigAdminAgent := NewBinary("./launcher-config-admin-agent", Targets64ExceptMacOS)
	launcherArchive.AddAuxiliarBinary(launcherConfigAdminAgent)
	// Full archive
	fullServerArchive := serverArchive.CloneWithFilesPrefix(`server`)
	fullLauncherArchive := launcherArchive.CloneWithFilesPrefix(`launcher`)
	fullBattleServerManager := battleServerManagerArchive.CloneWithFilesPrefix(`battle-server-manager`)
	fullArchive := NewMergedArchive("full", fullServerArchive, fullLauncherArchive, fullBattleServerManager)
	fullArchive.RemoveFiles("LICENSE")
	fullArchive.AddDocFiles("docs", nil, nil, "LICENSE", "README.md")
	GenerateConfig(serverArchive, battleServerManagerArchive, launcherArchive, fullArchive)
}
