package main

import (
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

func main() {
	commonLogger.Initialize(os.Stdout)
	lock := &fileLock.PidLock{}
	exitCode := common.ErrSuccess
	if err := lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'agent' if it is running in your task manager.")
		exitCode = common.ErrPidLock
		return
	}
	common.ChdirToExe()
	var steamProcess bool
	if runtime.GOOS == "windows" {
		steamProcess, _ = strconv.ParseBool(os.Args[1])
	} else {
		steamProcess = true
	}
	commonLogger.Printf("Steam process: %v\n", steamProcess)
	var xboxProcess bool
	if runtime.GOOS == "windows" {
		xboxProcess, _ = strconv.ParseBool(os.Args[2])
	}
	commonLogger.Printf("Xbox process: %v\n", xboxProcess)
	gameId := os.Args[5]
	commonLogger.Printf("Game ID: %s\n", gameId)
	serverExe := os.Args[3]
	commonLogger.Printf("Server executable: %s\n", serverExe)
	var broadcastBattleServer bool
	if runtime.GOOS == "windows" && (gameId != common.GameAoM && gameId != common.GameAoE4) {
		broadcastBattleServer, _ = strconv.ParseBool(os.Args[4])
	}
	commonLogger.Printf("Broadcast LAN Battle-Server: %v\n", broadcastBattleServer)
	battleServerExe := os.Args[6]
	commonLogger.Printf("Battle Server Manager executable: %s\n", battleServerExe)
	battleServerRegion := os.Args[7]
	commonLogger.Printf("Battle Server region: %s\n", battleServerRegion)
	logRoot := os.Args[8]
	commonLogger.Printf("Log root: %s\n", logRoot)
	if logRoot != "-" {
		internal.Initialize(logRoot)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			commonLogger.Println("Received terminate signal, shutting down...")
			exitCode = common.ErrSignal
			defer func() {
				if err := lock.Unlock(); err != nil {
					commonLogger.Printf("Failed to unlock: %v\n", err)
				}
				commonLogger.Printf("Exit code: %d\n", exitCode)
				os.Exit(exitCode)
			}()
			_ = internal.Logger.Buffer("config_revert_end", func(writer io.Writer) {
				if !launcherCommon.ConfigRevert(gameId, logRoot, true, nil, func(options exec.Options) {
					if writer != nil {
						commonLogger.Println("run config revert", options.String())
					}
				}, nil) {
					commonLogger.Println("Failed to revert configuration")
				}
			})
			_ = internal.Logger.Buffer("revert_command_end", func(writer io.Writer) {
				if err := launcherCommon.RunRevertCommand(writer, func(options exec.Options) {
					if writer != nil {
						commonLogger.Println("run revert command", options.String())
					}
				}); err != nil {
					commonLogger.Printf("Failed to revert command: %v\n", err)
				}
			})
			if serverExe != "-" {
				commonLogger.Println("Killing server...")
				if err := commonProcess.Kill(serverExe); err != nil {
					commonLogger.Printf("Failed to kill server: %v\n", serverExe)
				}
				if battleServerExe != "-" && battleServerRegion != "-" {
					commonLogger.Println("Shutting down battle-server...")
					_ = internal.Logger.Buffer("battle-server-manager_remove", func(writer io.Writer) {
						if result := launcherCommon.RemoveBattleServerRegion(battleServerExe, gameId, battleServerRegion, writer, func(options exec.Options) {
							if writer != nil {
								commonLogger.Println("run battle-server-manager", options.String())
							}
						}); !result.Success() {
							commonLogger.Println("Failed to shut down battle-server.")
							if result.Err != nil {
								commonLogger.Println(result.Err)
							}
							if result.ExitCode != common.ErrSuccess {
								commonLogger.Printf("Exit code: %d\n", result.ExitCode)
							}
						}
					})
				}
			}
		}
	}()
	watch.Watch(
		gameId,
		logRoot,
		steamProcess,
		xboxProcess,
		serverExe,
		broadcastBattleServer,
		battleServerExe,
		battleServerRegion,
		&exitCode,
	)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
