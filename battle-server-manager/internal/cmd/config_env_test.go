package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/v2"
	bsManager "github.com/luskaner/ageLANServer/common/cmd/bsManager"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/spf13/pflag"
)

// Regression: the Docker image ships AGELANSERVER_BATTLE_SERVER_MANAGER_*
// environment variables whose keys the env provider lowercases. Defaults and
// config files must therefore use lowercase keys too, or the env layer can
// never override them (ports 27012/27112/27212 silently ignored).
func TestStartConfigEnvOverridesLowercaseKeys(t *testing.T) {
	t.Setenv("AGELANSERVER_BATTLE_SERVER_MANAGER_PORTS_BS", "27012")
	t.Setenv("AGELANSERVER_BATTLE_SERVER_MANAGER_CERTSPATH", "/app/server/resources/certificates")

	k := koanf.New(".")
	defaults := map[string]any{
		"region":               "auto",
		"name":                 "auto",
		"host":                 "auto",
		"certspath":            "auto",
		"executable.path":      "auto",
		"executable.extraargs": []string{},
		"ports.bs":             0,
		"ports.websocket":      0,
		"ports.outofband":      0,
	}
	fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
	if _, err := common.LoadKoanfLayers(k, defaults, nil, toml.Parser(), fs, nil, executables.BattleServerManager); err != nil {
		t.Fatal(err)
	}

	if got := k.Int("ports.bs"); got != 27012 {
		t.Fatalf("env must override ports.bs, got %d", got)
	}
	if got := k.String("certspath"); got != "/app/server/resources/certificates" {
		t.Fatalf("env must override certspath, got %q", got)
	}
}

func TestInitConfigReadsLowercaseToml(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.age2.toml")
	const contents = `region = 'eu'
name = 'auto'
host = '192.168.1.10'
certspath = 'auto'
[executable]
path = 'auto'
extraargs = []
[ports]
bs = 21001
websocket = 21002
outofband = 21003
`
	if err := os.WriteFile(cfgFile, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}

	values, flags := bsManager.StartFlagSet(configPaths)
	if err := flags.Parse([]string{"--game", "age2", "--gameConfig", cfgFile}); err != nil {
		t.Fatal(err)
	}
	c := initConfig(flags, values)
	if c.Region != "eu" || c.Host != "192.168.1.10" {
		t.Fatalf("scalars wrong: region=%q host=%q", c.Region, c.Host)
	}
	if c.Ports.Bs != 21001 || c.Ports.WebSocket != 21002 || c.Ports.OutOfBand != 21003 {
		t.Fatalf("ports wrong: %+v", c.Ports)
	}
}
