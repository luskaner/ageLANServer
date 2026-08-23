package logger

import (
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/game"
	gameCert "github.com/luskaner/ageLANServer/common/game/cert"
	"github.com/luskaner/ageLANServer/common/hosts"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/userData"
)

var processesLog = []string{executables.LauncherAgent, executables.LauncherConfigAdminAgent}
var LogEnabled bool

// sessionState holds the per-run configuration set by root.go and read by
// WriteFileLog and its callees (which may run concurrently via the signal
// goroutine). Access must go through the exported getters/setters.
type sessionState struct {
	mu                     sync.RWMutex
	Cacert                 *gameCert.CA
	BasePath               string
	MacOsExclusiveMappings bool
}

var state sessionState

func SetCacert(ca *gameCert.CA) {
	state.mu.Lock()
	state.Cacert = ca
	state.mu.Unlock()
}

func SetBasePath(p string) {
	state.mu.Lock()
	state.BasePath = p
	state.mu.Unlock()
}

func SetMacOsExclusiveMappings(v bool) {
	state.mu.Lock()
	state.MacOsExclusiveMappings = v
	state.mu.Unlock()
}

func GetBasePath() string {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.BasePath
}
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
		state.mu.RLock()
		allHosts := common.AllHosts(gameId, state.MacOsExclusiveMappings)
		cacert := state.Cacert
		basePath := state.BasePath
		state.mu.RUnlock()

		if err := writeLog(gameId, "Auxiliar processes status", writeProcessesStatus); err != nil {
			log.Println(err)
		}
		if err := writeLog(gameId, "Relevant installed certificates", func(_ string) error {
			return writeUserPcCertificateInfo(allHosts, gameId)
		}); err != nil {
			commonLogger.Println(err)
		}
		if cacert != nil {
			if err := writeLog(gameId, "Relevant game installed certificates", func(_ string) error {
				return writeGameCertificateInfo(cacert, allHosts, gameId)
			}); err != nil {
				commonLogger.Println(err)
			}
		}
		if gameId != game.AoE1 {
			if err := writeLog(gameId, "Metadata folders", func(_ string) error {
				return writeMetadataInfo(basePath, gameId)
			}); err != nil {
				log.Println(err)
			}
		}
		if err := writeLog(gameId, "Profile folders", func(_ string) error {
			return writeProfilesInfo(basePath, gameId)
		}); err != nil {
			log.Println(err)
		}
		if err := writeLog(gameId, "Relevant host entries", func(_ string) error {
			return writeHostInfo(allHosts)
		}); err != nil {
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
		path := executables.NativeFileName(false, processName)
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

func writeHostInfo(allHosts []string) error {
	f, err := hosts.OpenMain()
	if err != nil {
		return fmt.Errorf("error opening hosts: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	var lines []hosts.Line
	if err, _, lines = hosts.GetAllLines(f); err == nil {
		addedSomeEntry := false
		allHostsSet := mapset.NewThreadUnsafeSet[string](allHosts...)
		for _, line := range lines {
			hsts := line.Hosts()
			hostsSet := mapset.NewThreadUnsafeSet[string]()
			for _, host := range hsts {
				hostsSet.Add(string(host))
			}
			if hostsSet.ContainsAnyElement(allHostsSet) {
				commonLogger.Printf("%s", line.String())
				addedSomeEntry = true
			}
		}
		if !addedSomeEntry {
			commonLogger.Println("No matchings.")
		}
	} else {
		return fmt.Errorf("error reading hosts: %w", err)
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

func writeCertificateInfo(certs []*x509.Certificate, allHosts []string) error {
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

func writeGameCertificateInfo(cacert *gameCert.CA, allHosts []string, _ string) error {
	files := []string{cacert.TmpPath(), cacert.BackupPath(), cacert.OriginalPath()}
	for _, file := range files {
		str := filepath.Base(file) + ": "
		_, _, certs, err := common.ReadFromFile(file)
		if err != nil {
			commonLogger.Println(str + err.Error())
			continue
		}
		commonLogger.Println(str)
		if err := writeCertificateInfo(certs, allHosts); err != nil {
			commonLogger.Println(err.Error())
		}
	}
	return nil
}

func writePcCertificateInfo(userStore bool, allHosts []string) error {
	commonLogger.Printf("Certificates of user %t\n", userStore)
	certs, err := cert.EnumCertificates(userStore)
	if err != nil {
		return fmt.Errorf("failed to enumerate %t certificates: %v", userStore, err)
	}
	return writeCertificateInfo(certs, allHosts)
}

func writeUserPcCertificateInfo(allHosts []string, _ string) error {
	localErr := writePcCertificateInfo(false, allHosts)
	userErr := writePcCertificateInfo(true, allHosts)
	return errors.Join(localErr, userErr)
}

func writeMetadataInfo(basePath string, gameId string) error {
	if basePath == "" {
		commonLogger.Println("Unknown")
		return nil
	}
	if err, metadatas := userData.NewPath(basePath, gameId).Metadatas(); err != nil {
		return err
	} else {
		writeDataInfo(metadatas)
		return nil
	}
}

func writeProfilesInfo(basePath string, gameId string) error {
	if basePath == "" {
		commonLogger.Println("Unknown")
		return nil
	}
	if err, metadatas := userData.NewPath(basePath, gameId).Profiles(); err != nil {
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
		counter[data.Type()]++
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
