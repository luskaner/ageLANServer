package hosts

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/fileLock"
	launcherCommonHosts "github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-common/hosts/text"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

func getExistingHosts(f *os.File) (err error, encType int, existingHosts mapset.Set[launcherCommonHosts.Host]) {
	var lines []launcherCommonHosts.Line
	err, encType, lines = launcherCommonHosts.GetAllLines(f)
	if err != nil {
		return
	}
	existingHosts = mapset.NewThreadUnsafeSet[launcherCommonHosts.Host]()
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

func restoreBackup(backFile *os.File, mainLock *fileLock.Lock) error {
	return launcherCommonHosts.UpdateHosts(mainLock, func(f *os.File) error {
		_, err := f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		var written int64
		written, err = io.Copy(f, backFile)
		if err != nil {
			return err
		}
		return f.Truncate(written)
	}, FlushDns)
}

func restoreInPlace(mainLock *fileLock.Lock) error {
	err, encType, existingHosts := getExistingHosts(mainLock.File)
	if err != nil {
		return err
	}
	var enc encoding.Encoding
	enc, err = text.GetEncoding(encType)
	if existingHosts.IsEmpty() {
		return nil
	}
	_, err = mainLock.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return launcherCommonHosts.UpdateHosts(mainLock, func(f *os.File) error {
		var lines []string
		var line string

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		decodingReader := transform.NewReader(f, enc.NewDecoder())
		scanner := bufio.NewScanner(decodingReader)
		for scanner.Scan() {
			line = scanner.Text()
			ok, parsedLine := launcherCommonHosts.ParseLine(line)
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
		var linesJoinedEncoded string
		linesJoinedEncoded, err = enc.NewEncoder().String(linesJoined)
		if err != nil {
			return err
		}
		var n int
		n, err = f.WriteString(linesJoinedEncoded)
		if err != nil {
			return err
		}

		err = f.Truncate(int64(n))
		if err != nil {
			return err
		}
		return nil
	}, FlushDns)
}

func RemoveHosts() error {
	var backExists bool
	backLock, err := launcherCommonHosts.OpenLockedBackup(os.O_RDWR)
	if err == nil {
		backExists = true
	}
	mainLock, err := launcherCommonHosts.OpenLockedMain(os.O_RDWR)
	var createdMain bool
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if backExists {
				mainLock, err = launcherCommonHosts.OpenLockedMain(os.O_RDWR | os.O_CREATE)
				if err != nil {
					return err
				}
				createdMain = true
			} else {
				return nil
			}
		} else {
			return err
		}
	}
	if backExists {
		defer func() {
			backLockName := backLock.File.Name()
			_ = backLock.Unlock()
			_ = os.Remove(backLockName)
		}()
		var doRestoreBackup bool
		if createdMain {
			doRestoreBackup = true
		} else if backInfo, err := backLock.File.Stat(); err != nil {
			return err
		} else if mainInfo, err := mainLock.File.Stat(); err != nil {
			return err
		} else {
			doRestoreBackup = mainInfo.ModTime().Before(backInfo.ModTime().Add(1 * time.Second))
		}
		if doRestoreBackup {
			return restoreBackup(backLock.File, mainLock)
		}
	}
	return restoreInPlace(mainLock)
}
