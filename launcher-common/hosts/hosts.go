package hosts

import (
	"bufio"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"io"
	"os"
	"regexp"
	"strings"
)

const HostEndMarking = "#" + common.Name
const WindowsLineEnding = "\r\n"

var mappingRegExp = regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+(?P<host>\S+)`)

func CreateTemp() (*os.File, error) {
	return os.CreateTemp("", common.Name+"_host_*.txt")
}

func HostMappings() map[string]string {
	mappings := make(map[string]string)
	if cmd.MapIP != nil {
		for _, host := range common.AllHosts() {
			mappings[host] = cmd.MapIP.String()
		}
	}
	if cmd.MapCDN {
		mappings[launcherCommon.CDNDomain] = launcherCommon.CDNIP
	}
	return mappings
}

func mapping(line string) (string, string) {
	uncommentedLine := LineWithoutComment(line)
	matches := mappingRegExp.FindStringSubmatch(uncommentedLine)
	if matches == nil {
		return "", ""
	}
	return matches[1], matches[2]
}

func missingIpMappings(mappings *map[string]string, hostFilePath string) (err error, f *os.File) {
	err, f = OpenHostsFile(hostFilePath)
	if err != nil {
		return
	}
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineIp, lineHost := mapping(line)
		if ip, ok := (*mappings)[lineHost]; ok && ip == lineIp {
			delete(*mappings, lineHost)
		}
	}
	if err := scanner.Err(); err != nil {
		_ = UnlockFile(f)
		_ = f.Close()
	}
	return
}

func LineWithoutComment(line string) string {
	return strings.Split(line, "#")[0]
}

func OpenHostsFile(hostFilePath string) (err error, f *os.File) {
	f, err = os.OpenFile(
		hostFilePath,
		os.O_RDWR,
		0666,
	)
	if err == nil {
		err = LockFile(f)
		if err != nil {
			_ = f.Close()
			f = nil
		}
	}
	return
}

func UpdateHosts(hostsFile *os.File, updater func(file *os.File) error, flushFn func() (result *exec.Result)) error {
	closed := false
	var tmp *os.File = nil

	closeHostsFile := func() {
		_ = UnlockFile(hostsFile)
		_ = hostsFile.Sync()
		_ = hostsFile.Close()
		closed = true
	}

	removeTmpFile := func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		tmp = nil
	}

	defer func() {
		if !closed {
			closeHostsFile()
		}
		if tmp != nil {
			removeTmpFile()
		}
	}()
	var err error
	tmp, err = CreateTemp()
	if err != nil {
		return err
	}

	_, err = io.Copy(tmp, hostsFile)

	if err != nil {
		return err
	}

	if err = updater(tmp); err == nil {
		err = hostsFile.Truncate(0)
		if err != nil {
			return err
		}

		_, err = hostsFile.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		_, err = io.Copy(hostsFile, tmp)
		if err != nil {
			return err
		}
		removeTmpFile()
		closeHostsFile()
		if flushFn != nil {
			_ = flushFn()
		}
		return nil
	}

	return err
}

func AddHosts(hostFilePath string, lineEnding string, flushFn func() (result *exec.Result)) (ok bool, err error) {
	var hostsFile *os.File
	mappings := HostMappings()
	err, hostsFile = missingIpMappings(&mappings, hostFilePath)
	if err != nil {
		return
	}

	if len(mappings) == 0 {
		_ = UnlockFile(hostsFile)
		_ = hostsFile.Close()
		ok = true
		return
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		_ = UnlockFile(hostsFile)
		_ = hostsFile.Close()
		return
	}

	err = UpdateHosts(hostsFile, func(f *os.File) error {
		for hostname, ip := range mappings {
			_, err = f.WriteString(fmt.Sprintf("%s%s\t%s\t%s", lineEnding, ip, hostname, HostEndMarking))
			if err != nil {
				return err
			}
		}
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		return nil
	}, flushFn)
	ok = err == nil
	return
}
