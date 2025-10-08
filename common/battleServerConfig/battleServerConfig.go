package battleServerConfig

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/process"
	"github.com/pelletier/go-toml/v2"
)

type BaseConfig struct {
	// Cannot be an UUID as it can be confused for a LAN one.
	Region string
	// Only used for common.GameAoE2
	Name          string
	IPv4          string
	BsPort        int
	WebSocketPort int
	// Used for all except common.GameAoE1
	OutOfBandPort int
}

type Config struct {
	BaseConfig `toml:",inline"`
	PID        uint32
	index      int `toml:"-"`
}

func ParseFileName(fileName string) (int, error) {
	index := common.BaseNameNoExt(fileName)
	return strconv.Atoi(index)
}

func Folder(gameId string) string {
	return filepath.Join(os.TempDir(), common.Name, "battle-servers", gameId)
}

func Configs(gameId string, onlyValid bool) (configs []Config, err error) {
	folder := Folder(gameId)
	var entries []os.DirEntry
	entries, err = os.ReadDir(folder)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
			return
		}
		err = fmt.Errorf("error while reading battle servers config directory \"%s\": %v", folder, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var index int
		index, err = ParseFileName(entry.Name())
		if err != nil {
			continue
		}
		var data []byte
		path := filepath.Join(folder, entry.Name())
		data, err = os.ReadFile(path)
		if err != nil {
			err = fmt.Errorf("error while reading battle server config file \"%s\": %v", entry.Name(), err)
			return
		}
		var config Config
		if err = toml.Unmarshal(data, &config); err != nil {
			err = fmt.Errorf("error while parsing battle server config file \"%s\": %v", entry.Name(), err)
			return
		}
		if !onlyValid || config.Validate() {
			config.index = index
			configs = append(configs, config)
		}
	}
	return
}

func (c Config) Validate() bool {
	if c.Region == "" || c.PID == 0 || c.IPv4 == "" || c.BsPort == 0 || c.WebSocketPort == 0 {
		return false
	}
	proc, err := process.FindProcess(int(c.PID))
	if err != nil || proc == nil {
		return false
	}
	ports := []int{c.BsPort, c.WebSocketPort}
	if c.OutOfBandPort != 0 {
		ports = append(ports, c.OutOfBandPort)
	}
	IPv4 := c.IPv4
	if IPv4 == "auto" {
		IPv4 = netip.IPv4Unspecified().String()
	}
	for _, port := range ports {
		target := net.JoinHostPort(IPv4, strconv.Itoa(port))
		conn, portErr := net.DialTimeout("tcp4", target, 100*time.Millisecond)
		if portErr != nil {
			return false
		}
		_ = conn.Close()
	}
	return true
}

func (c Config) Path() string {
	return Name(c.index)
}

func Name(index int) string {
	return fmt.Sprintf("%d.toml", index)
}
