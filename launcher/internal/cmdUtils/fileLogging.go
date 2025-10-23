package cmdUtils

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/spf13/viper"
)

var processesLog = []string{common.LauncherAgent, common.LauncherConfigAdminAgent}
var logFile *os.File
var allHosts []string

func OpenFileLog(gameId string) error {
	if viper.GetBool("Config.Log") {
		t := time.Now()
		var err error
		logFile, err = os.OpenFile(
			filepath.Join("logs", gameId, fmt.Sprintf("%d-%02d-%02dT%02d-%02d-%02d.txt", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())),
			os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			return err
		}
		log.SetOutput(logFile)
		log.SetFlags(0)
	}
	return nil
}

func CloseFileLog() {
	if viper.GetBool("Config.Log") && logFile != nil {
		_ = logFile.Close()
	}
}

func logPrefix(name string) {
	log.SetPrefix("|" + strings.ToUpper(name) + "| ")
}

func WriteFileLog(gameId string, name string) {
	if viper.GetBool("Config.Log") {
		logPrefix(name)
		allHosts = common.AllHosts(gameId)
		if err := writeLog(gameId, "Auxiliar processes Status", writeProcessesStatus); err != nil {
			log.Println(err)
		}
		if err := writeLog(gameId, "Relevant installed certificates", writeCertificateInfo); err != nil {
			log.Println(err)
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

func LogPrintf(name string, format string, a ...any) {
	if viper.GetBool("Config.Log") {
		logPrefix(name)
		log.Printf(format, a...)
	}
}

func LogPrint(name string, a ...any) {
	if viper.GetBool("Config.Log") {
		logPrefix(name)
		log.Print(a...)
	}
}

func LogPrintln(name string, a ...any) {
	if viper.GetBool("Config.Log") {
		logPrefix(name)
		log.Println(a...)
	}
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
		log.Println(str)
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
				log.Printf(line.String())
				addedSomeEntry = true
			}
		}
		if !addedSomeEntry {
			log.Println("No matchings.")
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
		log.Println("No arguments.")
	} else {
		log.Println(strings.Join(flags, " "))
	}
	return nil
}

func writeRevertConfigArgs(_ string) error {
	err, flags := launcherCommon.RevertConfigStore.Load()
	if err != nil {
		return fmt.Errorf("error reading revert config args: %w", err)
	}
	if len(flags) == 0 {
		log.Println("No arguments.")
	} else {
		log.Println(strings.Join(flags, " "))
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
		log.Println("No certificates.")
	} else {
		for _, crt := range matchingCerts {
			dnsGames := "No DNS Names."
			if len(crt.DNSNames) > 0 {
				dnsGames = strings.Join(crt.DNSNames, ", ")
			}
			log.Printf("%s: %s\n", crt.Subject.CommonName, dnsGames)
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

func writeSection(section string) {
	log.Printf("# %s\n", section)
}

func writeLog(gameId string, name string, logger func(gameId string) error) error {
	nameCaps := strings.ToUpper(name)
	log.Printf("========== %s ==========\n", nameCaps)
	err := logger(gameId)
	if err != nil {
		return fmt.Errorf("failed to write log content text: %v", err)
	}
	return nil
}
