package cmdUtils

import (
	"fmt"
	"io"
	"net"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/hosts"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/server"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) MapHosts(gameId string, ip string, macOsExclusiveMappings bool, canMap bool, customHostFile bool) (exitCode int) {
	var mapIP bool
	if !customHostFile {
		for _, domain := range common.AllHosts(gameId, macOsExclusiveMappings) {
			if !common.Matches(ip, domain) {
				if !canMap {
					logger.Println("serverStart is false and canAddHost is false but 'server' does not match " + domain + ". You should have added the host ip mapping to it in the hosts file (or just set canAddHost to true).")
					exitCode = internal.ErrConfigIpMap
					return
				}
				mapIP = true
			} else if err := server.CheckConnectionFromServer(domain, true, nil); err != nil {
				logger.Println("serverStart is false and host matches. " + domain + " must be reachable. Review the host is reachable via this domain to TCP port 443 (HTTPS).")
				logger.Printf("Error: %s\n", err.Error())
				exitCode = internal.ErrServerUnreachable
				return
			}
		}
	} else {
		mapIP = true
	}
	if mapIP {
		var str string
		if customHostFile {
			hostFileLock, err := hosts.CreateTemp()
			if err != nil {
				return internal.ErrConfigIpMapAdd
			}
			tmpName := hostFileLock.File.Name()
			c.hostFilePath, _ = filepath.Abs(tmpName)
			str += fmt.Sprintf("Saving hosts to '%s' file", tmpName)
			if err = hostFileLock.Unlock(); err != nil {
				return internal.ErrConfigIpMapAdd
			}
		} else {
			str += "Adding hosts to hosts file"
		}
		logger.Println(str + "...")
		var err error
		if err = commonLogger.FileLogger.Buffer("config_setup_hosts", func(writer io.Writer) {
			cfgSetupOpts := executor.NewConfigSetupOptions()
			cfgSetupOpts.Out = writer
			cfgSetupOpts.OptionsFn = func(options exec.Options) {
				commonLogger.Println("run config setup for hosts", options.String())
			}
			cfgSetupOpts.GameId = gameId
			cfgSetupOpts.MapIp = net.ParseIP(ip)
			cfgSetupOpts.MacOsExclusiveMappings = macOsExclusiveMappings
			cfgSetupOpts.HostFilePath = c.hostFilePath
			cfgSetupOpts.AgentEndOnError = !c.RequiresConfigRevert()
			if result := cfgSetupOpts.RunSetUp(); !result.Success() {
				logger.Println("Failed to add hosts.")
				if result.Err != nil {
					logger.Println("Error message: " + result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
				}
				exitCode = internal.ErrConfigIpMapAdd
			} else if customHostFile {
				if parsedIP := net.ParseIP(ip); parsedIP != nil {
					mappings := hosts.Mappings(gameId, parsedIP, macOsExclusiveMappings)
					for hostToCache, ipToCache := range mappings {
						common.CacheMapping(string(hostToCache), ipToCache.String())
					}
				} else {
					exitCode = internal.ErrConfigIpMapAdd
				}
			}
		}); err != nil {
			return common.ErrFileLog
		}
	}
	return
}
