package internal

import "github.com/gookit/ini/v2"

const Domain = "aoe-api.worldsedgelink.com"

type ServerConfig struct {
	Start               string
	Executable          string
	Host                string
	Stop                string
	WaitForProcessStart int
}

type ClientConfig struct {
	Executable               string
	WaitForProcessStart      int
	CheckProcessRunningEvery int
}

type Config struct {
	CanAddHost          bool
	CanTrustCertificate bool
	IsolateMetadata     bool
	Server              ServerConfig
	Client              ClientConfig
}

func ReadConfig() Config {
	cfg := ini.New()
	err := cfg.LoadFiles("config/config.ini")
	if err != nil {
		panic(err)
	}
	var config Config
	err = cfg.MapStruct("", &config)
	if err != nil {
		panic(err)
	}
	return config
}
