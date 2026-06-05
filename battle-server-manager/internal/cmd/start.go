package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"battle-server-manager/internal/cmdUtils/executor"
	"battle-server-manager/internal/cmdUtils/resolver"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/paths"
	"github.com/luskaner/ageLANServer/common/process"
	"github.com/spf13/pflag"
)

var configPaths = []string{paths.ResourcesDir, "."}

func runStart(args []string) error {
	values, flags := bsManager.StartFlagSet(configPaths)
	if err := flags.Parse(args); err != nil {
		return err
	}
	// validate required flags
	if values.GameId == "" {
		return errors.New("required flag 'game' not set")
	}

	cfg := initConfig(flags, values)
	gameIds := []string{values.GameId}
	games, err := cmdUtils.ParsedGameIds(&gameIds)
	if err != nil {
		commonLogger.Println(err.Error())
		os.Exit(internal.ErrGames)
	}
	values.GameId, _ = games.Pop()
	commonLogger.Println("Checking and resolving configuration...")
	isAdmin := commonExecutor.IsAdmin()
	if isAdmin {
		commonLogger.Println("Running as administrator, this is not needed and might cause issues.")
	}
	name := cfg.Name
	region := cfg.Region
	err, names, regions := cmdUtils.ExistingServers(values.GameId)
	if err != nil {
		commonLogger.Printf("could not get existing servers: %s\n", err.Error())
		os.Exit(internal.ErrReadConfig)
	}
	if !values.Force && !regions.IsEmpty() {
		if values.NoErrExisting {
			return nil
		}
		commonLogger.Println("a Battle Server is already running, use --force to start another one")
		os.Exit(internal.ErrAlreadyRunning)
	}
	if name == "auto" || region == "auto" {
		if name == "auto" {
			if names.ContainsOne("server") || regions.ContainsOne("server") {
				for i := 1; ; i++ {
					if currentName := fmt.Sprintf("Server (%d)", i); !names.ContainsOne(currentName) && !regions.ContainsOne(currentName) {
						name = currentName
						break
					}
				}
			} else {
				name = "Server"
			}
			commonLogger.Println("Auto-generated name:", name)
		}
		if region == "auto" {
			region = name
			commonLogger.Println("Auto-generated region:", region)
		}
	}
	if lowerRegion := strings.ToLower(region); names.ContainsOne(lowerRegion) || regions.ContainsOne(lowerRegion) {
		commonLogger.Printf("a Battle Server with the name/region '%s' already exists\n", region)
		os.Exit(internal.ErrAlreadyExists)
	}
	if lowerName := strings.ToLower(name); names.ContainsOne(lowerName) || regions.ContainsOne(lowerName) {
		commonLogger.Printf("a Battle Server with the name/region '%s' already exists\n", region)
		os.Exit(internal.ErrAlreadyExists)
	}
	host := cfg.Host
	var ip string
	if host != "auto" {
		ips := common.HostOrIpToIps(host)
		if len(ips) == 0 {
			commonLogger.Println("could not resolve host to an IP address")
			os.Exit(internal.ErrResolveHost)
		}
		for _, currentIP := range ips {
			if !net.ParseIP(currentIP).IsLoopback() {
				ip = currentIP
			}
		}
		if ip == "" {
			commonLogger.Println("ip not valid or could not resolve host to a suitable IP address")
			os.Exit(internal.ErrInvalidHost)
		}
		if ip != host {
			commonLogger.Println("Resolved host to IP address:", ip)
		}
	} else {
		ip = host
	}
	bsPort := cfg.Ports.Bs
	websocketPort := cfg.Ports.WebSocket
	outOfBandPort := -1
	if values.GameId != game.AoE1 {
		outOfBandPort = cfg.Ports.OutOfBand
	}
	if bsPort > 0 && !cmdUtils.Available(bsPort) {
		commonLogger.Printf("bs port %d is already in use\n", bsPort)
		os.Exit(internal.ErrBsPortInUse)
	}
	if websocketPort > 0 && !cmdUtils.Available(websocketPort) {
		commonLogger.Printf("websocket port %d is already in use\n", websocketPort)
		os.Exit(internal.ErrWsPortInUse)
	}
	if outOfBandPort > 0 && !cmdUtils.Available(outOfBandPort) {
		commonLogger.Printf("out of band port %d is already in use\n", outOfBandPort)
		os.Exit(internal.ErrOobPortInUse)
	}
	allPorts, err := cmdUtils.GeneratePortsAsNeeded([]int{bsPort, websocketPort, outOfBandPort})
	if err != nil {
		commonLogger.Printf("could not generate ports: %s\n", err)
		os.Exit(internal.ErrGenPorts)
	}
	if bsPort != allPorts[0] {
		commonLogger.Println("\tAuto-generated BsPort port:", allPorts[0])
	}
	if websocketPort != allPorts[1] {
		commonLogger.Println("\tAuto-generated WebSocketPort port:", allPorts[1])
	}
	if outOfBandPort != allPorts[2] {
		commonLogger.Println("\tAuto-generated Out Of Band Port:", allPorts[2])
	}
	resolvedCertFile, resolvedKeyFile, err := cmdUtils.ResolveSSLFilesPath(
		values.GameId,
		cfg.SSL,
	)
	if err != nil {
		commonLogger.Printf("could not resolve SSL files: %s\n", err)
		os.Exit(internal.ErrResolveSSLFiles)
	}
	resolvedPath, err := resolver.ResolvePath(values.GameId, cfg.Executable.Path)
	if err != nil {
		commonLogger.Printf("could not resolve path: %s\n", err)
		os.Exit(internal.ErrResolvePath)
	}
	extraArgs, err := common.ParseCommandArgsFromSlice(cfg.Executable.ExtraArgs, nil, true)
	if err != nil {
		commonLogger.Printf("could not parse extra args: %s\n", err)
		os.Exit(internal.ErrParseArgs)
	}
	var pid uint32
	pid, err = executor.ExecuteBattleServer(
		values.GameId,
		resolvedPath,
		region,
		name,
		allPorts,
		resolvedCertFile,
		resolvedKeyFile,
		extraArgs,
		values.HideWindow,
		values.LogRoot,
	)
	if err != nil {
		commonLogger.Printf("could not execute BattleServer: %s\n", err)
		os.Exit(internal.ErrStartBattleServer)
	}
	saveConfig := battleServer.Config{
		Base: battleServer.Base{
			Region:        region,
			Name:          name,
			IPv4:          ip,
			BsPort:        allPorts[0],
			WebSocketPort: allPorts[1],
		},
		PID: pid,
	}
	if allPorts[2] != -1 {
		saveConfig.OutOfBandPort = allPorts[2]
	}
	commonLogger.Println("Waiting up to 30s for the initialization to complete...")
	if !cmdUtils.WaitForBattleServerInit(saveConfig) {
		commonLogger.Printf("battle server initialization did not complete in time\n")
		if proc, localErr := process.FindProcess(int(saveConfig.PID)); localErr == nil && proc != nil {
			if localErr := process.KillProc(proc); localErr != nil {
				commonLogger.Println("Error: ", localErr)
			} else {
				commonLogger.Println("OK.")
			}
		} else if localErr != nil {
			commonLogger.Println("Could not find the process to kill: ", localErr)
		}
		os.Exit(internal.ErrInitBattleServer)
	}
	if err = cmdUtils.WriteConfig(values.GameId, saveConfig); err != nil {
		commonLogger.Printf("could not write config: %s\n", err)
		commonLogger.Println(err)
		commonLogger.Println("Stopping started Battle Server...")
		cmdUtils.Kill(saveConfig)
		os.Exit(internal.ErrConfigWrite)
	}
	return nil
}

func initConfig(fs *pflag.FlagSet, values *bsManager.StartValues) *internal.Configuration {
	k := koanf.New(".")
	defaults := map[string]any{
		"Region":               "auto",
		"Name":                 "auto",
		"Host":                 "auto",
		"Executable.Path":      "auto",
		"Executable.ExtraArgs": []string{},
		"Ports.Bs":             0,
		"Ports.WebSocket":      0,
		"Ports.OutOfBand":      0,
		"SSL.Auto":             true,
		"SSL.CertFile":         "",
		"SSL.KeyFile":          "",
	}

	var fileCandidates []string
	if values.GameCfgFile != "" {
		fileCandidates = append(fileCandidates, values.GameCfgFile)
	} else {
		for _, configPath := range configPaths {
			fileCandidates = append(fileCandidates, filepath.Join(configPath, fmt.Sprintf("config.%s.toml", values.GameId)))
		}
	}

	usedFile, err := common.LoadKoanfLayers(k, defaults, fileCandidates, toml.Parser(), fs, nil, executables.BattleServerManager)
	if err != nil {
		if fileErr, ok := errors.AsType[*common.KoanfFileLoadError](err); ok {
			commonLogger.Println("Error parsing config file:", fileErr.Path+":", fileErr.Err.Error())
		} else {
			commonLogger.Println("Error loading config:", err.Error())
		}
		os.Exit(common.ErrConfigParse)
	}
	if values.GameCfgFile != "" && usedFile == "" {
		commonLogger.Println("No config file found, using defaults.")
	}
	if usedFile != "" {
		commonLogger.Println("Using config file:", usedFile)
		if values.LogRoot != "" {
			data, _ := os.ReadFile(usedFile)
			commonLogger.PrefixPrintln("config", string(data))
		}
	}

	var c internal.Configuration
	if err := k.Unmarshal("", &c); err != nil {
		commonLogger.Printf("unable to decode configuration: %v\n", err)
		os.Exit(common.ErrConfigParse)
	}
	return &c
}
