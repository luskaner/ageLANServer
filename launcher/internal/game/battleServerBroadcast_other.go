//go:build !windows

package game

import (
	"github.com/luskaner/ageLANServer/common/config/shared"
	"net"
)

func RebroadcastIPs(filterIPs []net.IP, filterInterfaces []shared.Interface) (IPs []net.IP) {
	return
}
