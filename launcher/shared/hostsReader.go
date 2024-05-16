package shared

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var mappingRegExp = regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+(?P<Host>\S+)`)
var hostRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+(?P<Host>\S+)`)

func resolveHost(host string) *string {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil
	}
	return &addrs[0]
}

func ResolveHost(host string) string {
	if net.ParseIP(host) == nil {
		return *resolveHost(host)
	}
	return host
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

func Matches(addr1 string, addr2 string) bool {
	if addr1 == "0.0.0.0" {
		for _, ip := range getLocalIps() {
			if Matches(ip, addr2) {
				return true
			}
		}
		return false
	}
	domainIp := resolveHost(addr2)
	if domainIp == nil {
		return false
	}
	return addr1 == *domainIp
}

func HostsFile() string {
	return filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
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

func Host(line string) string {
	uncommentedLine := lineWithoutComment(line)
	matches := hostRegExp.FindStringSubmatch(uncommentedLine)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func HostExists(host string) bool {
	f, err := os.Open(HostsFile())
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
		lineHost := Host(line)
		if lineHost == host {
			return true
		}
	}
	return false
}

func MappingExists(ip string, host string) bool {
	f, err := os.Open(HostsFile())
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
		if lineIp == ip && lineHost == host {
			return true
		}
	}
	return false
}
