package shared

import (
	"bufio"
	"common"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"os"
	"shared/executor"
	"strings"
)

func AddHosts(ips mapset.Set[string]) bool {
	missingIps := MissingIpMappings(ips, common.Domain)
	if missingIps.IsEmpty() {
		return true
	}
	p := HostsFile()
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Sync()
		_ = f.Close()
		flushDns()
	}()
	for ip := range missingIps.Iter() {
		_, _ = f.WriteString(fmt.Sprintf("\r\n%s\t%s\t#AoE2DELanServer", ip, common.Domain))
	}
	return true
}

func flushDns() bool {
	return executor.RunCustomExecutable("ipconfig", "/flushdns")
}

func RemoveHosts() bool {
	if !HostExists(common.Domain) {
		return true
	}
	f, err := os.OpenFile(HostsFile(), os.O_RDWR, 0644)
	if err != nil {
		return false
	}
	defer func(file *os.File) {
		_ = file.Sync()
		_ = file.Close()
		flushDns()
	}(f)

	var lines []string
	var line string

	for scanner := bufio.NewScanner(f); scanner.Scan(); line = scanner.Text() {
		lineHost := Host(line)
		if lineHost == common.Domain {
			continue
		}
		lines = append(lines, line)
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return false
	}

	linesJoined := strings.Join(lines[1:], "\r\n")
	_, err = f.WriteString(linesJoined)
	if err != nil {
		return false
	}

	err = f.Truncate(int64(len(linesJoined)))
	if err != nil {
		return false
	}

	return true
}
