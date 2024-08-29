//go:build !windows

package watch

import (
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
)

func waitForProcess(pid uint32) bool {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, unix.SIGCHLD)

	for {
		sig := <-sigc
		if sig == unix.SIGCHLD {
			var status unix.WaitStatus
			_, err := unix.Wait4(int(pid), &status, unix.WNOHANG, nil)
			if err != nil {
				return false
			}
			if status.Exited() {
				return true
			}
		}
	}
}
