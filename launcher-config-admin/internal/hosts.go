package internal

import (
	"bufio"
	"common"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sys/windows"
	"io"
	"launcher-common/executor"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var mappingRegExp = regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+(?P<host>\S+)`)
var hostRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+(?P<host>\S+)`)

const hostEndMarking = "#" + common.Name

func openHostsFile() (err error, f *os.File, lock *windows.Overlapped) {
	f, err = os.OpenFile(
		filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts"),
		os.O_RDWR,
		0666,
	)
	if f != nil {
		fileHandle := windows.Handle(f.Fd())
		lock = &windows.Overlapped{}
		err = windows.LockFileEx(
			fileHandle,
			windows.LOCKFILE_EXCLUSIVE_LOCK,
			0,
			1,
			0,
			lock,
		)
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

func unlockFile(file *os.File, lock *windows.Overlapped) error {
	fileHandle := windows.Handle(file.Fd())
	return windows.UnlockFileEx(fileHandle, 0, 1, 0, lock)
}

func hostExists(name string) (err error, exists bool, f *os.File, lock *windows.Overlapped) {
	err, f, lock = openHostsFile()
	if f == nil {
		return
	}

	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineHost := host(line)
		if lineHost == name {
			exists = true
			return
		}
	}
	return
}

func missingIpMappings(ips mapset.Set[string], host string) (err error, missingIps mapset.Set[string], f *os.File, lock *windows.Overlapped) {
	err, f, lock = openHostsFile()
	if f == nil {
		return
	}
	missingIps = ips.Clone()
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineIp, lineHost := mapping(line)
		if lineHost == host && missingIps.ContainsOne(lineIp) {
			missingIps.Remove(lineIp)
		}
	}
	return
}

func AddHosts(ips mapset.Set[string]) (ok bool, err error) {
	err, missingIps, hostsFile, lock := missingIpMappings(ips, common.Domain)
	if err != nil {
		return
	}

	if missingIps.IsEmpty() {
		_ = unlockFile(hostsFile, lock)
		_ = hostsFile.Close()
		ok = true
		return
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		_ = unlockFile(hostsFile, lock)
		_ = hostsFile.Close()
		return
	}

	err = updateHosts(hostsFile, *lock, func(f *os.File) error {
		for ip := range missingIps.Iter() {
			_, err = f.WriteString(fmt.Sprintf("\r\n%s\t%s\t%s", ip, common.Domain, hostEndMarking))
			if err != nil {
				return err
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

func flushDns() (result *executor.ExecResult) {
	result = executor.ExecOptions{File: "ipconfig", SpecialFile: true, UseWorkingPath: true, ExitCode: true, Wait: true, Args: []string{"/flushdns"}}.Exec()
	return
}

func updateHosts(hostsFile *os.File, lock windows.Overlapped, updater func(file *os.File) error) error {
	closed := false
	var tmp *os.File = nil

	closeHostsFile := func() {
		_ = unlockFile(hostsFile, &lock)
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
	tmp, err := os.CreateTemp("", common.Name+".*")
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

func RemoveHosts() error {
	err, exists, hostsFile, lock := hostExists(common.Domain)
	if !exists {
		if hostsFile != nil && lock != nil {
			_ = unlockFile(hostsFile, lock)
			_ = hostsFile.Close()
		}
		return nil
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		_ = unlockFile(hostsFile, lock)
		_ = hostsFile.Close()
		return err
	}

	return updateHosts(hostsFile, *lock, func(f *os.File) error {
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
				if lineHost != common.Domain {
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
