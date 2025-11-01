package serverKill

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/process"
)

func Do(path string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second,
	}
	ips := common.ResolveUnspecifiedIps()
	if len(ips) == 0 {
		return fmt.Errorf("could not resolve local IPs")
	}
	for _, ip := range ips {
		resp, err := client.Post(fmt.Sprintf("https://%s/shutdown", ip), "", nil)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
	}
	pid, proc, err := process.Process(path)
	if err != nil {
		return nil
	}
	wait := 2 * time.Second
	if process.WaitForProcess(proc, &wait) {
		return nil
	}
	if err = process.KillPidProc(pid, proc); err != nil {
		return err
	}
	return nil
}
