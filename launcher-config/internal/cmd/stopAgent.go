package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-config/internal/admin"
)

func runStopAgent(_ []string) (err error, exitCode int) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
		}
	}()
	commonLogger.Println("Stopping agent if needed...")
	_ = admin.StopAgentIfNeeded()
	return
}
