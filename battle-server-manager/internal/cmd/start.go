package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/paths"
	"github.com/luskaner/ageLANServer/common/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

var configPaths = []string{paths.ResourcesDir, "."}
var hideWindow bool
var gameId string
var force bool
var noErrExisting bool
var logRoot string

var (
	gameCfgFile string
	startCmd    = &cobra.Command{
		Use:   "start",
		Short: "Run Battle Server instances.",
		Long:  "Run Battle Server instances and setup configurations.",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := initConfig()
			gameIds := []string{gameId}
			games, err := cmdUtils.ParsedGameIds(&gameIds)
			if err != nil {
				commonLogger.Println(err.Error())
				os.Exit(internal.ErrGames)
			}
			gameId, _ := games.Pop()
			commonLogger.Println("Checking and resolving configuration...")
			isAdmin := commonExecutor.IsAdmin()
			if isAdmin {
				commonLogger.Println("Running as administrator, this is not needed and might cause issues.")
			}
			name := cfg.Name
			region := cfg.Region
			err, names, regions := cmdUtils.ExistingServers(gameId)
			if err != nil {
				commonLogger.Printf("could not get existing servers: %s\n", err.Error())
				os.Exit(internal.ErrReadConfig)
			}
			if !force && !regions.IsEmpty() {
				if noErrExisting {
					return
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
			if gameId != common.GameAoE1 {
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
				gameId,
				cfg.SSL,
			)
			if err != nil {
				commonLogger.Printf("could not resolve SSL files: %s\n", err)
				os.Exit(internal.ErrResolveSSLFiles)
			}
			resolvedPath, err := cmdUtils.ResolvePath(gameId, cfg.Executable.Path)
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
			pid, err = cmdUtils.ExecuteBattleServer(
				gameId,
				resolvedPath,
				region,
				name,
				allPorts,
				resolvedCertFile,
				resolvedKeyFile,
				extraArgs,
				hideWindow,
				logRoot,
			)
			if err != nil {
				commonLogger.Printf("could not execute BattleServer: %s\n", err)
				os.Exit(internal.ErrStartBattleServer)
			}
			saveConfig := battleServerConfig.Config{
				BaseConfig: battleServerConfig.BaseConfig{
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
			commonLogger.Println("Waiting up to 10s for the initialization to complete...")
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
			if err = cmdUtils.WriteConfig(gameId, saveConfig); err != nil {
				commonLogger.Printf("could not write config: %s\n", err)
				commonLogger.Println(err)
				commonLogger.Println("Stopping started Battle Server...")
				cmdUtils.Kill(saveConfig)
				os.Exit(internal.ErrConfigWrite)
			}
		},
	}
)

func InitStart() {
	cmd.LogRootCommand(startCmd.Flags(), &logRoot)
	cmd.GameVarCommand(startCmd.Flags(), &gameId)
	if err := startCmd.MarkFlagRequired("game"); err != nil {
		panic(err)
	}
	startCmd.Flags().StringVar(&gameCfgFile, "gameConfig", "", fmt.Sprintf(`Game config file (default config.game.toml in %s directories)`, strings.Join(configPaths, ", ")))
	startCmd.Flags().BoolVarP(&hideWindow, "hideWindow", "w", false, "Hide Battle Server window.")
	startCmd.Flags().BoolVarP(&force, "force", "f", false, "Force to start more than a single Battle Server per game.")
	startCmd.Flags().BoolVarP(&noErrExisting, "noErrExisting", "r", false, "When 'force' is true and one already exists, exit without error.")
	RootCmd.AddCommand(startCmd)
}

func initConfig() *internal.Configuration {
	v.SetDefault("Region", "auto")
	v.SetDefault("Name", "auto")
	v.SetDefault("Host", "auto")
	v.SetDefault("Executable.Path", "auto")
	v.SetDefault("SSL.Auto", true)
	for _, configPath := range configPaths {
		v.AddConfigPath(configPath)
	}
	v.SetConfigType("toml")
	if gameCfgFile != "" {
		v.SetConfigFile(gameCfgFile)
	} else {
		v.SetConfigName(fmt.Sprintf("config.%s", gameId))
	}
	if err := v.ReadInConfig(); err == nil {
		commonLogger.Println("Using config file:", v.ConfigFileUsed())
		data, _ := os.ReadFile(v.ConfigFileUsed())
		commonLogger.PrefixPrintln("config", string(data))
	} else {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			commonLogger.Println("No config file found, using defaults.")
		} else {
			commonLogger.Println("Error parsing config file:", v.ConfigFileUsed()+":", err.Error())
			os.Exit(common.ErrConfigParse)
		}
	}
	v.AutomaticEnv()
	var c *internal.Configuration
	if err := v.Unmarshal(&c); err != nil {
		commonLogger.Printf("unable to decode configuration: %v\n", err)
		os.Exit(common.ErrConfigParse)
	}
	return c
}
