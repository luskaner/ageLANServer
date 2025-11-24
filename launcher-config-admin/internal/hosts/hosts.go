package hosts

import (
	"bufio"
	"io"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	launcherCommonHosts "github.com/luskaner/ageLANServer/launcher-common/hosts"
)

func getExistingHosts() (err error, existingHosts mapset.Set[string], f *os.File) {
	var lines []launcherCommonHosts.Line
	err, lines, f = launcherCommonHosts.GetAllLines(os.O_RDWR)
	if err != nil {
		return
	}
	existingHosts = mapset.NewThreadUnsafeSet[string]()
	for _, line := range lines {
		if !line.Own() {
			continue
		}
		for _, lineHost := range line.Hosts() {
			existingHosts.Add(lineHost)
		}
	}
	return
}

func RemoveHosts() error {
	err, existingHosts, hostsFile := getExistingHosts()
	if err != nil {
		return err
	}
	if existingHosts.IsEmpty() {
		if hostsFile != nil && launcherCommonHosts.Lock != nil {
			launcherCommonHosts.CloseFile(hostsFile)
		}
		return nil
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		launcherCommonHosts.CloseFile(hostsFile)
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
			ok, parsedLine := launcherCommonHosts.Parse(line)
			lineToAdd := line
			if ok && parsedLine.Own() {
				lineToAdd = ""
			}
			if lineToAdd != "" {
				lines = append(lines, lineToAdd)
			}
		}

		if err = scanner.Err(); err != nil {
			return err
		}

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		linesJoined := strings.Join(lines, launcherCommonHosts.LineEnding)
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
