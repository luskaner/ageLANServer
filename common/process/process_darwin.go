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
	timeStr := strings.TrimSpace(string(output))
	if timeStr == "" {
		return 0, errors.New("empty process start time")
	}
	t, err := parseProcessStartTime(timeStr)
	if err != nil {
		return 0, err
	}
	return t.UnixNano(), nil
}

func parseProcessStartTime(timeStr string) (t time.Time, err error) {
	// lstart format examples:
	//   "Tue Dec  3 10:30:00 2024" (single-digit day, space-padded)
	//   "Tue Dec 24 10:30:00 2024" (double-digit day)
	// Note: lstart doesn't include timezone, so we parse in local timezone
	// since the process started in the local timezone of this system.
	t, err = time.ParseInLocation("Mon Jan _2 15:04:05 2006", timeStr, time.Local)
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
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Split only on first whitespace to handle process names with spaces
		// Format: "  PID COMM" where COMM may contain spaces
		fields := strings.SplitN(line, " ", 2)
		if len(fields) < 2 {
			continue
		}
		pidStr := strings.TrimSpace(fields[0])
		comm := strings.TrimSpace(fields[1])
		if pidStr == "" || comm == "" {
			continue
		}
		pid, err := strconv.ParseUint(pidStr, 10, 32)
		if err != nil {
			continue
		}
		cmdlineName := filepath.Base(comm)
		if slices.Contains(names, cmdlineName) {
			processesPid[cmdlineName] = uint32(pid)
		}
	}

	return processesPid
}
