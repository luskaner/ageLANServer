package hosts

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
)

const commentPrefix = "#"
const hostEndMarking = common.Name
const WindowsLineEnding = "\r\n"

type Line struct {
	ip      net.IP
	hosts   []string
	comment string
}

func (l Line) IP() net.IP {
	return l.ip
}

func (l Line) Hosts() []string {
	return l.hosts[:min(len(l.hosts), maxHostsPerLine)]
}

func (l Line) Own() bool {
	return l.comment == hostEndMarking
}

func (l Line) String() string {
	sb := strings.Builder{}
	sb.WriteString(l.ip.String())
	sb.WriteString("\t")
	sb.WriteString(strings.Join(l.hosts, ` `))
	if l.comment != "" {
		sb.WriteString(" ")
		sb.WriteString(commentPrefix)
		sb.WriteString(l.comment)
	}
	return sb.String()
}

func NewLine(ip net.IP, hosts []string) Line {
	return Line{
		ip:      ip,
		hosts:   hosts,
		comment: hostEndMarking,
	}
}

func Parse(line string) (ok bool, l Line) {
	l = Line{}
	split := strings.Split(line, "#")
	if len(split) > 1 {
		l.comment = split[1]
	}
	lineWithoutComment := split[0]
	if lineWithoutComment == "" {
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
	l.ip = net.ParseIP(lineWithoutCommentSep[0])
	l.hosts = make([]string, len(lineWithoutCommentSep)-1)
	for i, host := range lineWithoutCommentSep[1:min(len(lineWithoutCommentSep), maxHostsPerLine+1)] {
		if net.ParseIP(host) != nil {
			continue
		}
		l.hosts[i] = host
	}
	ok = true
	return
}

// HostMappings maps a host to an IP, but an IP can be mapped to multiple hosts.
type HostMappings map[string]net.IP

func (h *HostMappings) Set(host string, ip net.IP) {
	(*h)[strings.ToLower(host)] = ip
}

func (h *HostMappings) Get(host string) (ip net.IP, ok bool) {
	ip, ok = (*h)[strings.ToLower(host)]
	return
}

func (h *HostMappings) Delete(host string) {
	delete(*h, strings.ToLower(host))
}

func (h *HostMappings) String(lineEnding string) string {
	lines := make([]Line, 0)
	for host, ip := range *h {
		lines = append(
			lines,
			NewLine(ip[:], []string{host}),
		)
	}
	sb := strings.Builder{}
	for _, line := range lines {
		sb.WriteString(lineEnding)
		sb.WriteString(line.String())
	}
	return sb.String()
}

func CreateTemp() (*os.File, error) {
	return os.CreateTemp("", common.Name+"_host_*.txt")
}

func Mappings(gameId string) HostMappings {
	mappings := make(HostMappings)
	if cmd.MapIP != nil {
		for _, host := range common.AllHosts(gameId) {
			mappings.Set(host, cmd.MapIP)
		}
	}
	if cmd.MapCDN {
		ip := net.ParseIP(launcherCommon.CDNIP)
		mappings.Set(launcherCommon.CDNDomain, ip)
	}
	return mappings
}

func CloseFile(f *os.File) {
	_ = UnlockFile(f)
	_ = f.Close()
}

func GetAllLines(flag int) (err error, lines []Line, f *os.File) {
	err, f = openHostsFile(Path(), flag)
	if err != nil {
		return
	}
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		ok, parsedLine := Parse(line)
		if !ok {
			continue
		}
		lines = append(lines, parsedLine)
	}
	if err = scanner.Err(); err != nil {
		CloseFile(f)
	}
	return
}

func missingIpMappings(mappings *HostMappings, hostFilePath string) (err error, f *os.File) {
	err, f = openHostsFile(hostFilePath, os.O_RDWR)
	if err != nil {
		return
	}
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		ok, lineParsed := Parse(line)
		if !ok {
			continue
		}
		lineIp := lineParsed.IP()
		for _, lineHost := range lineParsed.Hosts() {
			if ip, ok := mappings.Get(lineHost); ok && ip.Equal(lineIp) {
				mappings.Delete(lineHost)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		CloseFile(f)
	}
	return
}

func openHostsFile(hostFilePath string, flag int) (err error, f *os.File) {
	f, err = os.OpenFile(
		hostFilePath,
		flag,
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

func AddHosts(gameId string, hostFilePath string, lineEnding string, flushFn func() (result *exec.Result)) (ok bool, err error) {
	if hostFilePath == "" {
		hostFilePath = Path()
	}
	if lineEnding == "" {
		lineEnding = LineEnding
	}
	var hostsFile *os.File
	mappings := Mappings(gameId)
	err, hostsFile = missingIpMappings(&mappings, hostFilePath)
	if err != nil {
		return
	}

	if len(mappings) == 0 {
		CloseFile(hostsFile)
		ok = true
		return
	}

	_, err = hostsFile.Seek(0, io.SeekStart)
	if err != nil {
		CloseFile(hostsFile)
		return
	}

	err = UpdateHosts(hostsFile, func(f *os.File) error {
		_, err = f.WriteString(mappings.String(lineEnding))
		if err != nil {
			return err
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
