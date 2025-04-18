package cmdUtils

import (
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"io"
	"net/http"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

func requiresMapCDN() bool {
	client := http.Client{Timeout: 3}
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

func (c *Config) MapHosts(host string, canMap bool, alreadySelectedIp bool) (errorCode int) {
	var mapCDN bool
	ips := mapset.NewThreadUnsafeSet[string]()
	if requiresMapCDN() {
		if !canMap {
			fmt.Println("canAddHost is false but CDN is required to be mapped. You should have added the", launcherCommon.CDNIP, "mapping to", launcherCommon.CDNDomain, "in the hosts file (or just set canAddHost to true).")
			errorCode = internal.ErrConfigCDNMap
			return
		}
		mapCDN = true
	}
	resolvedIps := mapset.NewThreadUnsafeSet[string]()
	for _, domain := range common.AllHosts() {
		if !launcherCommon.Matches(host, domain) {
			if !canMap {
				fmt.Println("serverStart is false and canAddHost is false but 'server' does not match " + domain + ". You should have added the host ip mapping to it in the hosts file (or just set canAddHost to true).")
				errorCode = internal.ErrConfigIpMap
				return
			} else {
				if alreadySelectedIp {
					ips.Add(host)
				} else {
					resolvedIps = resolvedIps.Union(launcherCommon.HostOrIpToIps(host))
				}
			}
		} else if !server.CheckConnectionFromServer(domain, true) {
			fmt.Println("serverStart is false and host matches. " + domain + " must be reachable. Review the host is reachable via this domain to TCP port 443 (HTTPS).")
			errorCode = internal.ErrServerUnreachable
			return
		}
	}
	if resolvedIps.Cardinality() > 0 {
		if ok, ip := SelectBestServerIp(resolvedIps.ToSlice()); !ok {
			fmt.Println("Failed to find a reachable IP for " + host + ".")
			errorCode = internal.ErrConfigIpMapFind
			return
		} else {
			ips.Add(ip)
		}
	}
	if !ips.IsEmpty() || mapCDN {
		fmt.Print("Adding host to hosts file")
		if !commonExecutor.IsAdmin() {
			fmt.Print(", authorize 'config-admin-agent' if needed")
		}
		fmt.Println("...")
		if result := executor.RunSetUp("", ips, nil, nil, false, false, mapCDN, true); !result.Success() {
			fmt.Println("Failed to add host.")
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			errorCode = internal.ErrConfigIpMapAdd
		} else {
			if !ips.IsEmpty() {
				c.MappedHosts()
			}
			if mapCDN {
				c.MappedCDN()
			}
		}
	}
	return
}
