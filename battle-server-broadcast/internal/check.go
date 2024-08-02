package internal

import "net"

func FlagsCheck(flags net.Flags) bool {
	return flags&net.FlagUp != 0 && flags&net.FlagLoopback == 0 && flags&net.FlagBroadcast != 0
}
