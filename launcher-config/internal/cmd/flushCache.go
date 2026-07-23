package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
	"github.com/luskaner/ageLANServer/launcher-config/internal/admin"
)

func runFlushCache(args []string) (err error, exitCode int) {
	flushCacheValues, flags := launcherCommonCmd.FlushCacheFlagSet()
	if err = flags.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
		}
	}()
	if flushCacheValues.LogRoot != "" {
		internal.Initialize(flushCacheValues.LogRoot)
	}
	if flushCacheValues.IPs || flushCacheValues.Certs {
		if executor.IsAdmin() {
			err, exitCode = admin.RunFlushCache(flushCacheValues.LogRoot, flushCacheValues.IPs, flushCacheValues.Certs)
			if err == nil && exitCode == common.ErrSuccess {
				commonLogger.Println("Successfully ran 'config-admin'")
			} else {
				if err != nil {
					commonLogger.Println("Received error:")
					commonLogger.Println(err)
				}
				if exitCode != common.ErrSuccess {
					commonLogger.Println("Received exit code:")
					commonLogger.Println(exitCode)
				}
				exitCode = internal.ErrAdminSetup
			}
		} else {
			agentStarted := admin.ConnectAgentIfNeeded() == nil
			if agentStarted {
				exitCode = internal.ErrAgentAlreadyStarted
				return
			}
			result := admin.StartAgent(flushCacheValues.IPs, flushCacheValues.Certs)
			if !result.Success() {
				commonLogger.Println("Failed to start 'config-admin-agent'")
				if result != nil {
					if result.Err != nil {
						commonLogger.Println(result.Err)
					}
					if result.ExitCode != common.ErrSuccess {
						commonLogger.Println(result.ExitCode)
					}
				}
				exitCode = internal.ErrStartAgent
			} else {
				agentStarted = admin.ConnectAgentIfNeededWithRetries()
				if !agentStarted {
					commonLogger.Println("Failed to connect to 'config-admin-agent' after starting it. Kill it using the task manager.")
					_ = admin.StopAgentIfNeeded()
					exitCode = internal.ErrStartAgentVerify
				}
			}
		}
	}
	return
}
