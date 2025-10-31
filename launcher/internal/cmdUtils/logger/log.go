package logger

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/spf13/viper"
)

var processesLog = []string{common.LauncherAgent, common.LauncherConfigAdminAgent}
var allHosts []string

func OpenMainFileLog(gameId string) error {
	if viper.GetBool("Config.Log") {
		err := commonLogger.NewOwnFileLogger("launcher", "", gameId, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteFileLog(gameId string, name string) {
	if commonLogger.FileLogger != nil {
		commonLogger.Prefix(name)
		allHosts = common.AllHosts(gameId)
		if err := writeLog(gameId, "Auxiliar processes status", writeProcessesStatus); err != nil {
			log.Println(err)
		}
		if err := writeLog(gameId, "Relevant installed certificates", writeCertificateInfo); err != nil {
			commonLogger.Println(err)
		}
		if err := writeLog(gameId, "Relevant host entries", writeHostInfo); err != nil {
			log.Println(err)
		}
		if err := writeLog(gameId, "Config revert arguments", writeRevertConfigArgs); err != nil {
			log.Println(err)
		}
		if err := writeLog(gameId, "Command revert arguments", writeRevertCommandArgs); err != nil {
			log.Println(err)
		}
	}
}

func Printf(format string, a ...any) {
	commonLogger.PrefixPrintf("main", format, a...)
	fmt.Printf(format, a...)
}

func Println(a ...any) {
	commonLogger.PrefixPrintln("main", a...)
	fmt.Println(a...)
}

func writeProcessesStatus(_ string) error {
	for _, processName := range processesLog {
		str := processName + ": "
		path := common.GetExeFileName(false, processName)
		_, _, err := process.Process(path)
		if err != nil {
			str += "dead/unknown"
		} else {
			str += "alive"
		}
		commonLogger.Println(str)
	}
	return nil
}

func writeHostInfo(_ string) error {
	if err, lines, f := hosts.GetAllLines(os.O_RDONLY); err != nil {
		return fmt.Errorf("error reading hosts: %w", err)
	} else {
		defer hosts.CloseFile(f)
		addedSomeEntry := false
		allHostsSet := mapset.NewThreadUnsafeSet[string](allHosts...)
		allHostsSet.Add(launcherCommon.CDNDomain)
		for _, line := range lines {
			hsts := line.Hosts()
			hostsSet := mapset.NewThreadUnsafeSet[string]()
			for _, host := range hsts {
				hostsSet.Add(strings.ToLower(host))
			}
			if hostsSet.ContainsAnyElement(allHostsSet) {
				commonLogger.Printf(line.String())
				addedSomeEntry = true
			}
		}
		if !addedSomeEntry {
			commonLogger.Println("No matchings.")
		}
	}
	return nil
}

func writeRevertCommandArgs(_ string) error {
	err, flags := launcherCommon.RevertCommandStore.Load()
	if err != nil {
		return fmt.Errorf("error reading revert command args: %w", err)
	}
	if len(flags) == 0 {
		commonLogger.Println("No arguments.")
	} else {
		commonLogger.Println(strings.Join(flags, " "))
	}
	return nil
}

func writeRevertConfigArgs(_ string) error {
	err, flags := launcherCommon.RevertConfigStore.Load()
	if err != nil {
		return fmt.Errorf("error reading revert config args: %w", err)
	}
	if len(flags) == 0 {
		commonLogger.Println("No arguments.")
	} else {
		commonLogger.Println(strings.Join(flags, " "))
	}
	return nil
}

func writeCertificateInfo(_ string) error {
	certs, err := cert.EnumCertificates(true)
	if err != nil {
		return fmt.Errorf("failed to enumerate certificates: %v", err)
	}
	matchingCerts := filterMatchingCerts(certs, allHosts)
	if len(matchingCerts) == 0 {
		commonLogger.Println("No certificates.")
	} else {
		for _, crt := range matchingCerts {
			dnsGames := "No DNS Names."
			if len(crt.DNSNames) > 0 {
				dnsGames = strings.Join(crt.DNSNames, ", ")
			}
			commonLogger.Printf("%s: %s\n", crt.Subject.CommonName, dnsGames)
		}
	}
	return nil
}

func matchPattern(pattern string, hosts []string) bool {
	for _, host := range hosts {
		if pattern == host {
			return true
		}
		if len(pattern) > 1 && pattern[0] == '*' && pattern[1] == '.' {
			suffix := pattern[1:]
			if len(host) <= len(suffix) {
				continue
			}
			if host[len(host)-len(suffix):] != suffix {
				continue
			}
			prefix := host[:len(host)-len(suffix)]
			if len(prefix) > 0 && !strings.Contains(prefix, ".") {
				return true
			}
		}
	}
	return false
}

func filterMatchingCerts(certs []*x509.Certificate, hosts []string) []*x509.Certificate {
	var matchingCerts []*x509.Certificate
	for _, crt := range certs {
		if strings.Contains(crt.Subject.CommonName, common.Name) {
			matchingCerts = append(matchingCerts, crt)
			goto nextCert
		} else if len(crt.DNSNames) > 0 {
			for _, san := range crt.DNSNames {
				if matchPattern(san, hosts) {
					matchingCerts = append(matchingCerts, crt)
					goto nextCert
				}
			}
		} else {
			if matchPattern(crt.Subject.CommonName, hosts) {
				matchingCerts = append(matchingCerts, crt)
				goto nextCert
			}
		}
	nextCert:
	}
	return matchingCerts
}

func writeLog(gameId string, name string, log func(gameId string) error) error {
	nameCaps := strings.ToUpper(name)
	commonLogger.Printf("========== %s ==========\n", nameCaps)
	err := log(gameId)
	if err != nil {
		return fmt.Errorf("failed to write log content text: %v", err)
	}
	return nil
}
