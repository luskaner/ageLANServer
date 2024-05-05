package internal

import (
	"bufio"
	"net"
	"os"
	"regexp"
	"strings"
)

var mappingRegExp = regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+(?P<host>\S+)`)
var hostRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+(?P<host>\S+)`)

func getIp(host string) *string {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil
	}
	return &addrs[0]
}

func getLocalIps() []string {
	addrs, err := net.InterfaceAddrs()
	var ips []string
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ips = append(ips, ipNet.IP.String())
		}
	}
	return ips
}

func MatchesDomain(address string) bool {
	if address == "0.0.0.0" {
		for _, ip := range getLocalIps() {
			if MatchesDomain(ip) {
				return true
			}
		}
		return false
	}
	var ip string
	if net.ParseIP(address) == nil {
		tmpIp := getIp(address)
		if tmpIp == nil {
			return false
		}
		ip = *tmpIp
	} else {
		ip = address
	}
	domainIp := getIp(Domain)
	if domainIp == nil {
		return false
	}
	return ip == *domainIp
}

func hostsFile() string {
	return os.Getenv("WINDIR") + `\System32\drivers\etc\hosts`
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

func mappingExists(ip string) bool {
	f, err := os.Open(hostsFile())
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Close()
	}()
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineIp, lineHost := mapping(line)
		if lineIp == ip && lineHost == Domain {
			return true
		}
	}
	return false
}

func hostExists() bool {
	f, err := os.Open(hostsFile())
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Close()
	}()
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineHost := host(line)
		if lineHost == Domain {
			return true
		}
	}
	return false
}

func AddHost(ip string) bool {
	if ip == "0.0.0.0" {
		ip = "127.0.0.1"
	}
	if mappingExists(ip) {
		return true
	}
	p := hostsFile()
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Sync()
		_ = f.Close()
		flushDns()
	}()
	_, err = f.WriteString("\r\n" + ip + "\t" + Domain + "\t#AoE2DELanServer\r\n")
	return err == nil
}

func flushDns() bool {
	return RunCustomExecutable("ipconfig", "/flushdns")
}

func RemoveHost() bool {
	if !hostExists() {
		return true
	}
	f, err := os.OpenFile(hostsFile(), os.O_RDWR, 0644)
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
		lineHost := host(line)
		if lineHost == Domain {
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
