package logger

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-common/userData"
)

var processesLog = []string{executables.LauncherAgent, executables.LauncherConfigAdminAgent}
var allHosts []string
var Cacert *cert.CA
var LogEnabled bool
var dataTypeToString = map[int]string{
	userData.TypeServer: "Own Backup",
	userData.TypeBackup: "Original Backup",
	userData.TypeActive: "Active",
}

func OpenMainFileLog(gameId string) error {
	if LogEnabled {
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
		if err := writeLog(gameId, "Relevant installed certificates", writePcCertificateInfo); err != nil {
			commonLogger.Println(err)
		}
		if Cacert != nil {
			if err := writeLog(gameId, "Relevant game installed certificates", writeGameCertificateInfo); err != nil {
				commonLogger.Println(err)
			}
		}
		if gameId != common.GameAoE1 {
			if err := writeLog(gameId, "Metadata folders", writeMetadataInfo); err != nil {
				log.Println(err)
			}
		}
		if err := writeLog(gameId, "Profile folders", writeProfilesInfo); err != nil {
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

func PrintFile(name string, path string) {
	if commonLogger.FileLogger != nil {
		data, _ := os.ReadFile(path)
		commonLogger.PrefixPrintln(name, string(data))
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
		path := executables.Filename(false, processName)
		_, proc, err := process.Process(path)
		if err != nil {
			str += "unknown"
		} else if proc == nil {
			str += "dead"
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
		for _, line := range lines {
			hsts := line.Hosts()
			hostsSet := mapset.NewThreadUnsafeSet[string]()
			for _, host := range hsts {
				hostsSet.Add(strings.ToLower(host))
			}
			if hostsSet.ContainsAnyElement(allHostsSet) {
				commonLogger.Printf("%s", line.String())
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

func writeCertificateInfo(certs []*x509.Certificate) error {
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

func writeGameCertificateInfo(_ string) error {
	files := []string{Cacert.TmpPath(), Cacert.BackupPath(), Cacert.OriginalPath()}
	for _, file := range files {
		str := filepath.Base(file) + ": "
		_, _, certs, err := cert.ReadFromFile(file)
		if err != nil {
			commonLogger.Println(str + err.Error())
			continue
		}
		commonLogger.Println(str)
		if err := writeCertificateInfo(certs); err != nil {
			commonLogger.Println(err.Error())
		}
	}
	return nil
}

func writePcCertificateInfo(_ string) error {
	certs, err := cert.EnumCertificates(true)
	if err != nil {
		return fmt.Errorf("failed to enumerate certificates: %v", err)
	}
	return writeCertificateInfo(certs)
}

func writeMetadataInfo(gameId string) error {
	if err, metadatas := userData.Metadatas(gameId); err != nil {
		return err
	} else {
		writeDataInfo(metadatas)
		return nil
	}
}

func writeProfilesInfo(gameId string) error {
	if err, metadatas := userData.Profiles(gameId); err != nil {
		return err
	} else {
		writeDataInfo(metadatas)
		return nil
	}
}

func writeDataInfo(datas mapset.Set[userData.Data]) {
	counter := map[int]int{}
	for typ := range dataTypeToString {
		counter[typ] = 0
	}
	for data := range datas.Iter() {
		counter[data.Type]++
	}
	for typ, count := range counter {
		commonLogger.Printf("%s: %d\n", dataTypeToString[typ], count)
	}
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
