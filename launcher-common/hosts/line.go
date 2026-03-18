package hosts

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/luskaner/ageLANServer/common"
)

const marking = common.Name

type Line struct {
	ip       net.IP
	hosts    []Host
	comments []string
}

func (l Line) IP() net.IP {
	return l.ip
}

func (l Line) Hosts() []Host {
	return l.hosts
}

func (l Line) Own() bool {
	if len(l.comments) == 0 {
		return false
	}
	return l.comments[len(l.comments)-1] == marking
}

func (l Line) OnlyComments() bool {
	return l.ip == nil && len(l.hosts) == 0
}

func (l Line) String() string {
	sb := strings.Builder{}
	if l.ip != nil {
		sb.WriteString(l.ip.String())
		sb.WriteString("\t")
		hostsStr := make([]string, len(l.hosts))
		for i, host := range l.hosts {
			hostsStr[i] = host.String()
		}
		sb.WriteString(strings.Join(hostsStr, ` `))
	}
	if len(l.comments) > 0 {
		sb.WriteString(commentMarker)
		sb.WriteString(strings.Join(l.comments, commentMarker))
	}
	return sb.String()
}

func (l Line) Commented() (ok bool, nl Line) {
	if l.OnlyComments() {
		return true, l
	}
	commented := fmt.Sprintf("%s%s", commentMarker, l.String())
	ok, _, nl = ParseLine(commented, true)
	return
}

func (l Line) Uncommented() (nl string) {
	if !l.OnlyComments() {
		return l.String()
	}
	sb := strings.Builder{}
	sb.WriteString(l.comments[0])
	if len(l.comments) > 1 {
		sb.WriteString(commentMarker)
		sb.WriteString(strings.Join(l.comments[1:], commentMarker))
	}
	return sb.String()
}

func (l Line) hasOwnMarking() bool {
	if len(l.comments) < 1 {
		return false
	}
	return l.comments[len(l.comments)-1] == marking
}

func (l Line) WithoutOwnMarking() Line {
	if !l.hasOwnMarking() {
		return l
	}
	l.comments = slices.Clone(l.comments[:len(l.comments)-1])
	return l
}

func (l Line) WithOwnMarking() Line {
	if l.hasOwnMarking() {
		return l
	}
	l.comments = slices.Clone(append(l.comments, marking))
	return l
}
