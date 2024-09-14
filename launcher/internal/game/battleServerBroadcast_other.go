//go:build !windows

package game

func RequiresBattleServerBroadcast() bool {
	return false
}
