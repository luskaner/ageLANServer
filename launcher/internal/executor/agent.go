package executor

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"net"
	"strconv"
	"strings"
)

func RunAgent(gameTitle common.GameTitle, steamProcess bool, xboxProcess bool, serverPid int, rebroadcastIPs []net.IP) (result *exec.Result) {
	rebroadcastIPsStr := make([]string, len(rebroadcastIPs))
	for i, ip := range rebroadcastIPs {
		rebroadcastIPsStr[i] = ip.String()
	}
	args := []string{
		strconv.FormatBool(steamProcess),
		strconv.FormatBool(xboxProcess),
		strconv.FormatInt(int64(serverPid), 10),
		strings.Join(rebroadcastIPsStr, ","),
		string(gameTitle),
	}
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}
