package shared

import (
	"bufio"
	mapset "github.com/deckarep/golang-set/v2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var mappingRegExp = regexp.MustCompile(`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+(?P<Host>\S+)`)
var hostRegExp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+(?P<Host>\S+)`)

func Matches(addr1 string, addr2 string) bool {
	addr2Ips := HostOrIpToIps(addr2)
	addr1Ips := HostOrIpToIps(addr1)
	return addr2Ips.Intersect(addr1Ips).Cardinality() > 0
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

func MissingIpMappings(ips mapset.Set[string], host string) mapset.Set[string] {
	missingIps := ips.Clone()
	f, err := os.Open(HostsFile())
	if err != nil {
		return nil
	}
	defer func() {
		_ = f.Close()
	}()
	var line string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		lineIp, lineHost := mapping(line)
		if lineHost == host && missingIps.ContainsOne(lineIp) {
			missingIps.Remove(lineIp)
		}
	}
	return missingIps
}
