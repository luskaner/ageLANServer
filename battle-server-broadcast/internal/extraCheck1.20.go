//go:build go1.20
// +build go1.20

package internal

import "net"

func FlagsExtraCheck(flags net.Flags) bool {
	return flags&net.FlagRunning != 0
}
