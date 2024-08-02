package cmd

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher/internal"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/executor"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/server"
)

func (c *Config) MapHosts(host string, canMap bool) (errorCode int) {
	if !launcherCommon.Matches(host, common.Domain) {
		if !canMap {
			fmt.Println("serverStart is false and canAddHost is false but server does not match "+common.Domain+". You should have added the host ip mapping to", common.Domain, "in the hosts file (or just set canAddHost to true).")
			errorCode = internal.ErrConfigIpMap
		} else {
			if commonExecutor.IsAdmin() {
				fmt.Println("Adding host to hosts file.")
			} else {
				fmt.Println(`Adding host to hosts file, accept any dialog from "launcher-config-admin" if it appears...`)
			}
			ips := launcherCommon.HostOrIpToIps(host)
			var ip string
			for curIp := range ips.Iter() {
				if server.LanServer(curIp, true) {
					ip = curIp
					break
				}
			}
			if ip == "" {
				fmt.Println("Failed to find a reachable IP for " + host + ".")
				errorCode = internal.ErrConfigIpMapFind
				return
			}
			if result := executor.RunSetUp(mapset.NewSet[string](ip), nil, nil, false, false, true); !result.Success() {
				fmt.Println("Failed to add host.")
				if result.Err != nil {
					fmt.Println("Error message: " + result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					fmt.Printf(`Exit code: %d. See documentation for "config" to check what it means.`+"\n", result.ExitCode)
				}
				errorCode = internal.ErrConfigIpMapAdd
			} else {
				c.MappedHosts()
			}
		}
	} else if !server.CheckConnectionFromServer(common.Domain, true) {
		fmt.Println("serverStart is false and host matches. " + common.Domain + " must be reachable. Review the host is reachable via this domain to TCP port 443 (HTTPS).")
		errorCode = internal.ErrServerUnreachable
	}
	return
}
