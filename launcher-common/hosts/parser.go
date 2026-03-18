package hosts

import (
	"net"
	"strings"

	"golang.org/x/net/idna"
)

func parseHost(host string) (ok bool, parsed Host) {
	var decoded string
	var err error
	if decoded, err = idna.Lookup.ToUnicode(host); err != nil {
		return
	}
	if net.ParseIP(decoded) == nil {
		ok = true
		parsed = Host(decoded)
	}
	return
}

func ParseLine(line string, ignoreLimit bool) (ok bool, overLimit bool, l Line) {
	l = Line{}
	var maxChars int
	if ignoreLimit {
		maxChars = len(line)
	} else {
		maxChars = maxCharsPerLine - len(LineEnding)
	}
	var usableLength int
	if actualLength := len(line); actualLength > maxChars {
		overLimit = true
		usableLength = maxChars
	} else {
		usableLength = actualLength
	}
	split := strings.SplitN(line[:usableLength], commentMarker, 2)
	if len(split) > 1 {
		l.comments = strings.Split(split[1], commentMarker)
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
	hosts := lineWithoutCommentSep[1:]
	var maxHosts int
	if ignoreLimit {
		maxHosts = len(hosts)
	} else {
		maxHosts = maxHostsPerLine
	}
	if actualLength := len(hosts); actualLength > maxHosts {
		overLimit = true
		usableLength = maxHosts
	} else {
		usableLength = actualLength
	}
	l.hosts = make([]Host, 0, usableLength)
	for _, host := range hosts[:usableLength] {
		if okParse, parsedHost := parseHost(host); okParse {
			l.hosts = append(l.hosts, parsedHost)
		}
	}
	ok = len(l.hosts) > 0
	return
}
