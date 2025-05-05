package hosts

import (
	"bufio"
	mapset "github.com/deckarep/golang-set/v2"
	launcherCommonHosts "github.com/luskaner/ageLANServer/launcher-common/hosts"
	"io"
	"os"
	"regexp"
	"strings"
)

var hostRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+(?P<host>\S+)`)

func host(line string) string {
	uncommentedLine := launcherCommonHosts.LineWithoutComment(line)
	matches := hostRegExp.FindStringSubmatch(uncommentedLine)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func getExistingHosts(hosts mapset.Set[string]) (err error, existingHosts mapset.Set[string], f *os.File) {
	err, f = launcherCommonHosts.OpenHostsFile(Path())
	if err != nil {
		return
	}
	existingHosts = mapset.NewThreadUnsafeSet[string]()
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineHost := host(line)
		if hosts.ContainsOne(lineHost) {
			existingHosts.Add(lineHost)
		}
	}
	if err = scanner.Err(); err != nil {
		_ = launcherCommonHosts.UnlockFile(f)
		_ = f.Close()
	}
	return
}

func RemoveHosts(hosts mapset.Set[string]) error {
	err, existingHosts, hostsFile := getExistingHosts(hosts)
	if err != nil {
		return err
	}
	if existingHosts.IsEmpty() {
		if hostsFile != nil && launcherCommonHosts.Lock != nil {
			_ = launcherCommonHosts.UnlockFile(hostsFile)
			_ = hostsFile.Close()
		}
		return nil
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		_ = launcherCommonHosts.UnlockFile(hostsFile)
		_ = hostsFile.Close()
		return err
	}

	return launcherCommonHosts.UpdateHosts(hostsFile, func(f *os.File) error {
		var lines []string
		var line string

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line = scanner.Text()
			addLine := false
			if !strings.HasSuffix(line, launcherCommonHosts.HostEndMarking) {
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

		if err = scanner.Err(); err != nil {
			return err
		}

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		linesJoined := strings.Join(lines, LineEnding)
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
	}, FlushDns)
}
