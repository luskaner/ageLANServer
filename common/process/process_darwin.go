package process

import (
	"os"
	"time"
)

func WaitForProcess(_ *os.Process, _ *time.Duration) bool {
	return true
}
