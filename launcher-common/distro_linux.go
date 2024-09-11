//go:build !darwin && !windows

package launcher_common

import (
	"bufio"
	"fmt"
	"github.com/google/shlex"
	"os"
	"strings"
)

var (
	distroID      string
	distroVersion string
	alreadyRead   bool
)

const ubuntuVersion = "24.04"
const steamOSMajorVersion = "3"

func read() {
	if alreadyRead {
		return
	}
	distroID, distroVersion = readOSReleaseVersion()
	alreadyRead = true
}

func readOSReleaseVersion() (id string, version string) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var tokens []string
	var lineTokens []string
	lines := bufio.NewScanner(file)

	for lines.Scan() {
		line := lines.Text()
		lineTokens, err = shlex.Split(line)
		if err != nil {
			continue
		}
		tokens = append(tokens, lineTokens...)
	}

	for _, token := range tokens {
		if strings.Contains(token, "=") {
			parts := strings.SplitN(token, "=", 2)
			key := strings.ToLower(parts[0])
			value := strings.Trim(strings.ToLower(parts[1]), `"`)
			switch key {
			case "id":
				id = value
			case "version_id":
				version = value
			}
		}
	}
	return
}

func Ubuntu() bool {
	id, _ := CheckDistro()
	return id == "ubuntu"
}

func SteamOS() bool {
	id, _ := CheckDistro()
	return id == "steamos"
}

func CheckDistro() (id string, err error) {
	read()
	if distroVersion == "" {
		err = fmt.Errorf("could not determine distro version")
		return
	}
	if distroID == "" {
		err = fmt.Errorf("could not determine distro identifier")
		return
	}
	switch distroID {
	case "ubuntu":
		if distroVersion == ubuntuVersion {
			id = "ubuntu"
			return
		} else {
			err = fmt.Errorf("unsupported Ubuntu version %s, only Ubuntu %s LTS is supported", distroVersion, ubuntuVersion)
		}
	case "steamos":
		if strings.HasPrefix(distroVersion, steamOSMajorVersion+".") {
			id = "steamos"
			return
		} else {
			err = fmt.Errorf("unsupported SteamOS version %s, only SteamOS %s is supported", distroVersion, steamOSMajorVersion)
		}
	default:
		err = fmt.Errorf("unsupported distro %s, only SteamOS %s and Ubuntu %s LTS are supported", steamOSMajorVersion, ubuntuVersion)
	}
	return
}
