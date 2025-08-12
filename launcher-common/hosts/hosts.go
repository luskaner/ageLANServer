package hosts

import (
	"bufio"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
)

const commentPrefix = "#"
const hostEndMarking = common.Name
const WindowsLineEnding = "\r\n"

type Line struct {
	ipAddr  netip.Addr
	hosts   []string
	comment string
}

func (l Line) IPAddr() netip.Addr {
	return l.ipAddr
}

func (l Line) Hosts() []string {
	return l.hosts[:min(len(l.hosts), maxHostsPerLine)]
}

func (l Line) Own() bool {
	return l.comment == hostEndMarking
}

func (l Line) String() string {
	sb := strings.Builder{}
	sb.WriteString(l.ipAddr.String())
	sb.WriteString("\t")
	sb.WriteString(strings.Join(l.hosts, ` `))
	if l.comment != "" {
		sb.WriteString(" ")
		sb.WriteString(commentPrefix)
		sb.WriteString(l.comment)
	}
	return sb.String()
}

func NewLine(ipAddr netip.Addr, hosts []string) Line {
	return Line{
		ipAddr:  ipAddr,
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
	if ip != nil {
		return
	}
	l.ipAddr, _ = netip.ParseAddr(lineWithoutCommentSep[0])
	l.hosts = make([]string, 0, len(lineWithoutCommentSep)-1)
	for i, host := range lineWithoutCommentSep[1:min(len(lineWithoutCommentSep), maxHostsPerLine+1)] {
		if net.ParseIP(host) != nil {
			continue
		}
		l.hosts[i] = host
	}
	ok = true
	return
}

// HostMappings maps a host to an IPAddr, but an IPAddr can be mapped to multiple hosts.
type HostMappings map[string]netip.Addr

func (h *HostMappings) String(lineEnding string) string {
	ipAddrToHosts := make(map[netip.Addr][]string)
	for host, ip := range *h {
		if _, ok := ipAddrToHosts[ip]; !ok {
			ipAddrToHosts[ip] = make([]string, 0)
		}
		ipAddrToHosts[ip] = append(ipAddrToHosts[ip], host)
	}
	lines := make([]Line, 0, len(ipAddrToHosts))
	for ipAddr, hosts := range ipAddrToHosts {
		lines = append(
			lines,
			NewLine(ipAddr, hosts),
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

func Mappings() HostMappings {
	mappings := make(HostMappings)
	if cmd.MapIPAddrValue.Addr.IsValid() {
		for _, host := range common.AllHosts() {
			mappings[host] = cmd.MapIPAddrValue.Addr
		}
	}
	if cmd.MapCDN {
		ipAddr, _ := netip.ParseAddr(launcherCommon.CDNIP)
		mappings[launcherCommon.CDNDomain] = ipAddr
	}
	return mappings
}

func missingIpMappings(mappings *HostMappings, hostFilePath string) (err error, f *os.File) {
	err, f = OpenHostsFile(hostFilePath)
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
		lineIp := lineParsed.IPAddr()
		for _, lineHost := range lineParsed.Hosts() {
			if ip, ok := (*mappings)[lineHost]; ok && ip == lineIp {
				delete(*mappings, lineHost)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		_ = UnlockFile(f)
		_ = f.Close()
	}
	return
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
	mappings := Mappings()
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
