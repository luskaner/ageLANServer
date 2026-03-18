package hosts

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/hosts/text"
	"golang.org/x/net/idna"
)

const commentPrefix = "#"
const hostEndMarking = common.Name
const WindowsLineEnding = "\r\n"

type Line struct {
	ip      net.IP
	hosts   []Host
	comment string
}

func (l Line) IP() net.IP {
	return l.ip
}

func (l Line) Hosts() []Host {
	return l.hosts
}

func (l Line) Own() bool {
	return l.comment == hostEndMarking
}

func (l Line) String() string {
	sb := strings.Builder{}
	var commentSep bool
	if l.ip != nil {
		sb.WriteString(l.ip.String())
		sb.WriteString("\t")
		hostsStr := make([]string, len(l.hosts))
		for i, host := range l.hosts {
			hostsStr[i] = host.String()
		}
		sb.WriteString(strings.Join(hostsStr, ` `))
		commentSep = true
	}
	if l.comment != "" {
		if commentSep {
			sb.WriteString(" ")
		}
		sb.WriteString(commentPrefix)
		sb.WriteString(l.comment)
	}
	return sb.String()
}

func NewOwnLine(ip net.IP, hosts []Host) Line {
	return Line{
		ip:      ip,
		hosts:   hosts,
		comment: hostEndMarking,
	}
}

type Host string

func ParseHost(host string) (ok bool, parsed Host) {
	var decoded string
	var err error
	if decoded, err = idna.Punycode.ToUnicode(host); err != nil {
		return
	}
	if net.ParseIP(decoded) == nil {
		ok = true
		parsed = Host(decoded)
	}
	return
}

func (h Host) String() string {
	norm, _ := idna.Punycode.ToASCII(string(h))
	return norm
}

func ParseLine(line string) (ok bool, l Line) {
	l = Line{}
	split := strings.SplitN(line, "#", 2)
	if len(split) > 1 {
		l.comment = split[1]
	}
	lineWithoutComment := split[0]
	if lineWithoutComment == "" {
		ok = true
		return
	}
	lineWithoutCommentSep := strings.Fields(lineWithoutComment)
	if len(lineWithoutCommentSep) < 2 {
		return
	}
	ip := net.ParseIP(lineWithoutCommentSep[0])
	if ip == nil {
		return
	}
	l.ip = ip
	l.hosts = make([]Host, 0, min(len(lineWithoutCommentSep)-1, maxHostsPerLine))
	for _, host := range lineWithoutCommentSep[1:min(len(lineWithoutCommentSep), maxHostsPerLine+1)] {
		if okParse, parsedHost := ParseHost(host); okParse {
			l.hosts = append(l.hosts, parsedHost)
		}
	}
	ok = true
	return
}

// HostMappings maps a host to an IP, but an IP can be mapped to multiple hosts.
type HostMappings map[Host]net.IP

func (h *HostMappings) Set(host Host, ip net.IP) {
	(*h)[host] = ip
}

func (h *HostMappings) Get(host Host) (ip net.IP, ok bool) {
	ip, ok = (*h)[host]
	return
}

func (h *HostMappings) Delete(host Host) {
	delete(*h, host)
}

func (h *HostMappings) String(lineEnding string) string {
	lines := make([]Line, len(*h))
	i := 0
	for host, ip := range *h {
		lines[i] = NewOwnLine(ip[:], []Host{host})
		i++
	}
	sb := strings.Builder{}
	for _, line := range lines {
		sb.WriteString(lineEnding)
		sb.WriteString(line.String())
	}
	sb.WriteString(lineEnding)
	return sb.String()
}

func CreateTemp() (lock *fileLock.Lock, err error) {
	var f *os.File
	f, err = os.CreateTemp("", common.Name+"_host_*.txt")
	if err != nil {
		return
	}
	lock = &fileLock.Lock{}
	if err = lock.Lock(f); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
	return
}
func OpenLockedBackup(flags int) (lock *fileLock.Lock, err error) {
	return openLockedHostsFile(filepath.Join(filepath.Dir(Path()), "hosts.bak"), flags)
}

func OpenLockedMain(flags int) (lock *fileLock.Lock, err error) {
	return openLockedHostsFile(Path(), flags)
}

func OpenMain() (f *os.File, err error) {
	return openHostsFile(Path(), os.O_RDONLY)
}

func Mappings(gameId string) HostMappings {
	mappings := make(HostMappings)
	if cmd.MapIP != nil {
		for _, host := range common.AllHosts(gameId) {
			mappings.Set(Host(host), cmd.MapIP)
		}
	}
	return mappings
}

func decodeAndGetScanner(f *os.File) (err error, enc int, scanner *bufio.Scanner) {
	var buf []byte
	buf, err = io.ReadAll(f)
	var str string
	str, enc, err = text.Decode(buf)
	if err != nil {
		return
	}
	scanner = bufio.NewScanner(strings.NewReader(str))
	return
}

func GetAllLines(f *os.File) (err error, encType int, lines []Line) {
	var scanner *bufio.Scanner
	err, encType, scanner = decodeAndGetScanner(f)
	if err != nil {
		return
	}
	addedHosts := mapset.NewThreadUnsafeSet[Host]()
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		ok, parsedLine := ParseLine(line)
		if !ok || parsedLine.IP() == nil {
			continue
		}
		var finalHosts []Host
		for _, host := range parsedLine.Hosts() {
			if !addedHosts.ContainsOne(host) {
				finalHosts = append(finalHosts, host)
				addedHosts.Add(host)
			}
		}
		if len(finalHosts) > 0 {
			lines = append(lines, NewOwnLine(parsedLine.ip, finalHosts))
		}
	}
	err = scanner.Err()
	return
}

func missingIpMappings(mappings *HostMappings, hostFile *os.File) (restLines []Line, err error) {
	var scanner *bufio.Scanner
	err, _, scanner = decodeAndGetScanner(hostFile)
	if err != nil {
		return
	}
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		ok, lineParsed := ParseLine(line)
		if !ok {
			continue
		}
		if lineParsed.IP() == nil {
			restLines = append(restLines, lineParsed)
			continue
		}
		lineIp := lineParsed.IP()
		hosts := lineParsed.Hosts()
		var indexesToAvoid []int
		for i, lineHost := range hosts {
			if ip, ok := mappings.Get(lineHost); ok && ip.Equal(lineIp) {
				mappings.Delete(lineHost)
				indexesToAvoid = append(indexesToAvoid, i)
			}
		}
		var keptHosts []Host
		for i := 0; i < len(hosts); i++ {
			if slices.Contains(indexesToAvoid, i) {
				continue
			}
			keptHosts = append(keptHosts, hosts[i])
		}
		if len(keptHosts) > 0 {
			restLines = append(restLines, Line{lineIp, keptHosts, ""})
		}
	}
	err = scanner.Err()
	return
}

func openHostsFile(hostFilePath string, flag int) (f *os.File, err error) {
	return os.OpenFile(
		hostFilePath,
		flag,
		0644,
	)
}

func openLockedHostsFile(hostFilePath string, flag int) (lock *fileLock.Lock, err error) {
	var f *os.File
	f, err = openHostsFile(hostFilePath, flag)
	if err == nil {
		lock = &fileLock.Lock{}
		if err = lock.Lock(f); err != nil {
			_ = f.Close()
			f = nil
		}
	}
	return
}

func UpdateHosts(hostsLock *fileLock.Lock, updater func(file *os.File) error, flushFn func() (result *exec.Result)) error {
	closed := false
	var tmpLock *fileLock.Lock = nil

	closeHostsFile := func() {
		_ = hostsLock.File.Sync()
		_ = hostsLock.Unlock()
		closed = true
	}

	removeTmpFile := func() {
		filePath := tmpLock.File.Name()
		_ = tmpLock.Unlock()
		_ = os.Remove(filePath)
		tmpLock = nil
	}

	defer func() {
		if !closed {
			closeHostsFile()
		}
		if tmpLock != nil {
			removeTmpFile()
		}
	}()
	var err error
	tmpLock, err = CreateTemp()
	if err != nil {
		return err
	}

	_, err = io.Copy(tmpLock.File, hostsLock.File)

	if err != nil {
		return err
	}

	if err = updater(tmpLock.File); err == nil {
		_, err = tmpLock.File.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		err = hostsLock.File.Truncate(0)
		if err != nil {
			return err
		}

		_, err = hostsLock.File.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		_, err = io.Copy(hostsLock.File, tmpLock.File)
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

func AddHosts(gameId string, hostFilePath string, lineEnding string, flushFn func() (result *exec.Result)) (ok bool, err error) {
	var systemHosts bool
	if hostFilePath == "" {
		systemHosts = true
		hostFilePath = Path()
	}
	if lineEnding == "" {
		lineEnding = LineEnding
	}
	var hostsFileLock *fileLock.Lock
	mappings := Mappings(gameId)
	var restLines []Line
	hostsFileLock, err = openLockedHostsFile(hostFilePath, os.O_RDWR)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			hostsFileLock, err = openLockedHostsFile(hostFilePath, os.O_RDWR|os.O_CREATE)
			if err != nil {
				return
			}
		}
	}
	restLines, err = missingIpMappings(&mappings, hostsFileLock.File)
	if err != nil {
		return
	}
	if len(mappings) == 0 {
		_ = hostsFileLock.Unlock()
		ok = true
		return
	}
	_, err = hostsFileLock.File.Seek(0, io.SeekStart)
	if err != nil {
		_ = hostsFileLock.Unlock()
		return
	}
	err = UpdateHosts(hostsFileLock, func(f *os.File) error {
		if systemHosts {
			var bakLock *fileLock.Lock
			bakLock, err = OpenLockedBackup(os.O_RDWR | os.O_CREATE | os.O_EXCL)
			if err != nil {
				return err
			}
			clearBakFunc := func() {
				filePath := bakLock.File.Name()
				_ = bakLock.Unlock()
				_ = os.Remove(filePath)
			}
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				clearBakFunc()
				return err
			}
			if _, err = io.Copy(bakLock.File, f); err != nil {
				clearBakFunc()
				return err
			}
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				clearBakFunc()
				return err
			}
		}
		for _, line := range restLines {
			if _, err = f.WriteString(line.String()); err != nil {
				return err
			}
			if _, err = f.WriteString(lineEnding); err != nil {
				return err
			}
		}
		_, err = f.WriteString(mappings.String(lineEnding))
		if err != nil {
			return err
		}
		return nil
	}, flushFn)
	ok = err == nil
	return
}
