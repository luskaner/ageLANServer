package hosts

import (
	"bufio"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"io"
	"os"
	"regexp"
	"strings"
)

var mappingRegExp = regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+(?P<host>\S+)`)
var hostRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+(?P<host>\S+)`)

const hostEndMarking = "#" + common.Name

func openHostsFile() (err error, f *os.File) {
	f, err = os.OpenFile(
		hostsPath(),
		os.O_RDWR,
		0666,
	)
	if f != nil {
		err = lockFile(f)
		if err != nil {
			_ = f.Close()
			f = nil
		}
	}
	return
}

func lineWithoutComment(line string) string {
	return strings.Split(line, "#")[0]
}

func mapping(line string) (string, string) {
	uncommentedLine := lineWithoutComment(line)
	matches := mappingRegExp.FindStringSubmatch(uncommentedLine)
	if matches == nil {
		return "", ""
	}
	return matches[1], matches[2]
}

func host(line string) string {
	uncommentedLine := lineWithoutComment(line)
	matches := hostRegExp.FindStringSubmatch(uncommentedLine)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func getExistingHosts(hosts mapset.Set[string]) (err error, existingHosts mapset.Set[string], f *os.File) {
	err, f = openHostsFile()
	if f == nil {
		return
	}
	existingHosts = mapset.NewSet[string]()
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineHost := host(line)
		if hosts.ContainsOne(lineHost) {
			existingHosts.Add(lineHost)
		}
	}
	return
}

func missingIpMappings(mappings *map[string]mapset.Set[string]) (err error, f *os.File) {
	err, f = openHostsFile()
	if f == nil {
		return
	}
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineIp, lineHost := mapping(line)
		if ips, ok := (*mappings)[lineHost]; ok && ips.ContainsOne(lineIp) {
			(*mappings)[lineHost].Remove(lineIp)
			if (*mappings)[lineHost].IsEmpty() {
				delete(*mappings, lineHost)
			}
		}
	}
	return
}

func AddHosts(mappings map[string]mapset.Set[string]) (ok bool, err error) {
	var hostsFile *os.File
	err, hostsFile = missingIpMappings(&mappings)
	if err != nil {
		return
	}

	if len(mappings) == 0 {
		_ = unlockFile(hostsFile)
		_ = hostsFile.Close()
		ok = true
		return
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		_ = unlockFile(hostsFile)
		_ = hostsFile.Close()
		return
	}

	err = updateHosts(hostsFile, func(f *os.File) error {
		for hostname, ips := range mappings {
			for ip := range ips.Iter() {
				_, err = f.WriteString(fmt.Sprintf("\r\n%s\t%s\t%s", ip, hostname, hostEndMarking))
				if err != nil {
					return err
				}
			}
		}
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		return nil
	})
	ok = err == nil
	return
}

func updateHosts(hostsFile *os.File, updater func(file *os.File) error) error {
	closed := false
	var tmp *os.File = nil

	closeHostsFile := func() {
		_ = unlockFile(hostsFile)
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
	tmp, err = os.CreateTemp("", common.Name+".*")
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
		_ = flushDns()
		return nil
	}

	return err
}

func RemoveHosts(hosts mapset.Set[string]) error {
	err, existingHosts, hostsFile := getExistingHosts(hosts)
	if existingHosts.IsEmpty() {
		if hostsFile != nil && lock != nil {
			_ = unlockFile(hostsFile)
			_ = hostsFile.Close()
		}
		return nil
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		_ = unlockFile(hostsFile)
		_ = hostsFile.Close()
		return err
	}

	return updateHosts(hostsFile, func(f *os.File) error {
		var lines []string
		var line string

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		for scanner := bufio.NewScanner(f); scanner.Scan(); line = scanner.Text() {
			addLine := false
			if !strings.HasSuffix(line, hostEndMarking) {
				addLine = true
			} else {
				lineHost := host(line)
				if !existingHosts.ContainsOne(lineHost) {
					addLine = true
				}
			}
			if addLine {
				lines = append(lines, line)
			}
		}

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		linesJoined := strings.Join(lines[1:], "\r\n")
		_, err = f.WriteString(linesJoined)
		if err != nil {
			return err
		}

		err = f.Truncate(int64(len(linesJoined)))
		if err != nil {
			return err
		}

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		return nil
	})
}
