package cmdUtils

import (
	"fmt"
	"io"
	"net"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func (c *Config) MapHosts(gameId string, ip string, canMap bool, customHostFile bool) (errorCode int) {
	var mapIP bool
	ips := mapset.NewThreadUnsafeSet[string]()
	if !customHostFile {
		for _, domain := range common.AllHosts(gameId) {
			if !common.Matches(ip, domain) {
				if !canMap {
					logger.Println("serverStart is false and canAddHost is false but 'server' does not match " + domain + ". You should have added the host ip mapping to it in the hosts file (or just set canAddHost to true).")
					errorCode = internal.ErrConfigIpMap
					return
				}

				mapIP = true
			} else if !server.CheckConnectionFromServer(domain, true, nil) {
				logger.Println("serverStart is false and host matches. " + domain + " must be reachable. Review the host is reachable via this domain to TCP port 443 (HTTPS).")
				errorCode = internal.ErrServerUnreachable
				return
			}
		}
	} else {
		mapIP = true
	}
	if mapIP {
		ips.Add(ip)
	}
	if !ips.IsEmpty() {
		var str string
		if customHostFile {
			hostFile, err := hosts.CreateTemp()
			if err != nil {
				return internal.ErrConfigIpMapAdd
			}
			if err = hostFile.Close(); err != nil {
				return internal.ErrConfigIpMapAdd
			}
			c.hostFilePath, _ = filepath.Abs(hostFile.Name())
			str += fmt.Sprintf("Saving hosts to '%s' file", hostFile.Name())
		} else {
			str += "Adding hosts to hosts file"
			if !commonExecutor.IsAdmin() {
				str += ", authorize 'config-admin-agent' if needed"
			}
		}
		logger.Println(str + "...")
		var err error
		if err = commonLogger.FileLogger.Buffer("config_setup_hosts", func(writer io.Writer) {
			if result := executor.RunSetUp(gameId, ips, nil, nil, nil, false, false, true, c.hostFilePath, "", "", writer, func(options exec.Options) {
				commonLogger.Println("run config setup for hosts", options.String())
			}); !result.Success() {
				logger.Println("Failed to add hosts.")
				if result.Err != nil {
					logger.Println("Error message: " + result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
				}
				errorCode = internal.ErrConfigIpMapAdd
			} else if customHostFile {
				if parsedIP := net.ParseIP(ip); parsedIP != nil {
					cmd.MapIP = parsedIP
				}
				mappings := hosts.Mappings(gameId)
				for hostToCache, ipToCache := range mappings {
					common.CacheMapping(hostToCache, ipToCache.String())
				}
			}
		}); err != nil {
			return common.ErrFileLog
		}
	}
	return
}
