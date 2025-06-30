package cmdUtils

import (
	"encoding/json"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"io"
	"net"
	"net/http"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

func requiresMapCDN() bool {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://%s/aoe/rl-server-status.json", launcherCommon.CDNDomain))
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	var result map[string]interface{}
	if err = json.Unmarshal(body, &result); err != nil {
		return false
	}
	var ok bool
	var startTime string
	if startTime, ok = result["start-time"].(string); !ok {
		return false
	}
	var endTime string
	if endTime, ok = result["end-time"].(string); !ok {
		return false
	}
	var startTimeParsed time.Time
	if startTimeParsed, err = time.Parse(timeLayout, startTime); err != nil {
		return false
	}
	var endTimeParsed time.Time
	if endTimeParsed, err = time.Parse(timeLayout, endTime); err != nil {
		return false
	}
	if endTimeParsed.Before(startTimeParsed) {
		return false
	}
	now := time.Now().UTC()
	if now.After(startTimeParsed) && now.Before(endTimeParsed) {
		return true
	}
	// Check time window is within 8 hours of now
	upperLimit := now.Add(8 * time.Hour)
	return (startTimeParsed.Before(upperLimit) && startTimeParsed.After(now)) || (endTimeParsed.Before(upperLimit) && endTimeParsed.After(now)) || (startTimeParsed.Before(now) && endTimeParsed.After(upperLimit))
}

func (c *Config) MapHosts(ip string, store string) (errorCode int) {
	var mapCDN bool
	var mapIP bool
	var ipToMap string
	if store != "tmp" {
		if requiresMapCDN() {
			if store == "" {
				printer.Println(
					printer.Error,
					printer.TS("CanAddHost", printer.OptionStyle),
					printer.T(" is "),
					printer.TS("false", printer.LiteralStyle),
					printer.T(" but CDN is required to be mapped. You should have added the mapping yourself."),
				)
				errorCode = internal.ErrConfigCDNMap
				return
			}
			mapCDN = true
		}
		for _, domain := range common.AllHosts() {
			if !launcherCommon.Matches(ip, domain) {
				if store == "" {
					printer.Println(
						printer.Error,
						printer.TS("Server.Start", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("false", printer.LiteralStyle),
						printer.T(" but "),
						printer.TS("server", printer.LiteralStyle),
						printer.T(" does not match "),
						printer.TS(domain, printer.LiteralStyle),
						printer.T("."),
					)
					errorCode = internal.ErrConfigIpMap
					return
				} else {
					mapIP = true
				}
			} else if !server.CheckConnectionFromServer(domain, true) {
				printer.Println(
					printer.Error,
					printer.TS("Server.Start", printer.OptionStyle),
					printer.T(" is "),
					printer.TS("false", printer.LiteralStyle),
					printer.T(" and host matches. "),
					printer.TS(domain, printer.LiteralStyle),
					printer.T(" must be reachable."),
				)
				errorCode = internal.ErrServerUnreachable
				return
			}
		}
	} else {
		mapIP = true
	}
	if mapIP {
		ipToMap = ip
	}
	if ipToMap != "" || mapCDN {
		if store == "tmp" {
			hostFile, err := hosts.CreateTemp()
			if err != nil {
				return internal.ErrConfigIpMapAdd
			}
			if err = hostFile.Close(); err != nil {
				return internal.ErrConfigIpMapAdd
			}
			c.SetHostFilePath(hostFile.Name())
			printer.Println(
				printer.Configuration,
				printer.T("Storing hosts to temporary file: "),
				printer.TS(hostFile.Name(), printer.FilePathStyle),
				printer.T("..."),
			)
		} else {
			addHostsStyledTexts := []*printer.StyledText{printer.T("Adding hosts to hosts file")}
			if !commonExecutor.IsAdmin() {
				addHostsStyledTexts = append(
					addHostsStyledTexts,
					printer.T(", authorize "),
					printer.TS("config-admin", printer.ComponentStyle),
					printer.T(" if needed"),
				)
			}
			addHostsStyledTexts = append(addHostsStyledTexts, printer.T("..."))
			fmt.Print(printer.Gen(printer.Configuration, "", addHostsStyledTexts...))
		}
		if result := executor.RunSetUp(&executor.RunSetUpOptions{HostFilePath: c.hostFilePath, MapIp: ipToMap, MapCDN: mapCDN, ExitAgentOnError: true}); !result.Success() {
			printer.PrintFailedResultError(result)
			errorCode = internal.ErrConfigIpMapAdd
		} else if store == "tmp" {
			cmd.MapCDN = true
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				cmd.MapIP = parsedIP
			}
			mappings := hosts.HostMappings()
			for hostToCache, ipToCache := range mappings {
				launcherCommon.CacheMapping(hostToCache, ipToCache)
			}
		}
		printer.PrintSucceeded()
	}
	return
}
