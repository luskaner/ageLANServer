package process

import (
	"errors"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func GetProcessStartTime(pid int) (int64, error) {
	// Use ps command to get process start time (lstart format)
	// This avoids brittle hardcoded struct offsets that vary between macOS versions/architectures
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=").Output()
	if err != nil {
		return 0, err
	}
	// lstart output format: "Mon Jan  2 15:04:05 2006"
	// We parse it and convert to nanoseconds since epoch
	timeStr := strings.TrimSpace(string(output))
	if timeStr == "" {
		return 0, errors.New("empty process start time")
	}
	// Parse the time string
	t, err := parseProcessStartTime(timeStr)
	if err != nil {
		return 0, err
	}
	return t.UnixNano(), nil
}

func parseProcessStartTime(timeStr string) (t time.Time, err error) {
	// lstart format: "Tue Dec 24 10:30:00 2024"
	t, err = time.Parse("Mon Jan _2 15:04:05 2006", timeStr)
	return
}

// ProcessesPID returns a map of process names to their PIDs.
// Note: If multiple processes share the same name, only one PID is stored per name.
func ProcessesPID(names []string) map[string]uint32 {
	processesPid := make(map[string]uint32)

	output, err := exec.Command("ps", "-axo", "pid,comm").Output()
	if err != nil {
		return processesPid
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			continue
		}
		cmdlineName := filepath.Base(fields[1])
		if slices.Contains(names, cmdlineName) {
			processesPid[cmdlineName] = uint32(pid)
		}
	}

	return processesPid
}
