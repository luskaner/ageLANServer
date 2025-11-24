package internal

import (
	"os"
	"os/signal"
	"syscall"
)

var StopSignal chan os.Signal

func InitializeStopSignal() {
	StopSignal = make(chan os.Signal)
	signal.Notify(StopSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
}
