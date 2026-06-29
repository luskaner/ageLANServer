package cmd

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
	"github.com/spf13/pflag"
)

var Version string
var values *agent.Values

func Execute() (err error, exitCode int) {
	var singleFs *cmd.SingleFlagSet
	values, singleFs = agent.SingleFlagSet(Version, runRoot)
	return singleFs.Execute()
}

func runRoot(_ *pflag.FlagSet) (err error, exitCode int) {
	commonLogger.Initialize(os.Stdout)
	lock := &fileLock.PidLock{}
	if err = lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'agent' if it is running in your task manager.")
		exitCode = common.ErrPidLock
		return
	}
	common.ChdirToExe()
	if values.LogRoot != "" && values.BaseDataPath != "-" {
		internal.Initialize(values.LogRoot)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			commonLogger.Println("Received terminate signal, shutting down...")
			exitCode = common.ErrSignal
			defer func() {
				if err = lock.Unlock(); err != nil {
					commonLogger.Printf("Failed to unlock: %v\n", err)
				}
				commonLogger.Printf("Exit code: %d\n", exitCode)
			}()
			_ = internal.Logger.Buffer("config_revert_end", func(writer io.Writer) {
				if !launcherCommon.ConfigRevert(values.GameId, values.LogRoot, true, nil, func(options exec.Options) {
					if writer != nil {
						commonLogger.Println("run config revert", options.String())
					}
				}, nil) {
					commonLogger.Println("Failed to revert configuration")
				}
			})
			_ = internal.Logger.Buffer("revert_command_end", func(writer io.Writer) {
				if err = launcherCommon.RunRevertCommand(writer, func(options exec.Options) {
					if writer != nil {
						commonLogger.Println("run revert command", options.String())
					}
				}); err != nil {
					commonLogger.Printf("Failed to revert command: %v\n", err)
				}
			})
			if values.ServerExecutable != "" {
				commonLogger.Println("Killing server...")
				if err = commonProcess.Kill(values.ServerExecutable); err != nil {
					commonLogger.Printf("Failed to kill server: %v\n", values.ServerExecutable)
				}
				if values.BattleServerManagerExecutable != "-" && values.BattleServerRegion != "-" {
					commonLogger.Println("Shutting down battle-server...")
					_ = internal.Logger.Buffer("battle-server-manager_remove", func(writer io.Writer) {
						if result := launcherCommon.RemoveBattleServerRegion(values.BattleServerManagerExecutable, values.GameId, values.BattleServerRegion, writer, func(options exec.Options) {
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
		values,
		&exitCode,
	)
	_ = lock.Unlock()
	return
}
