package internal

import "github.com/gookit/ini/v2"

type ServerConfig struct {
	Start      string
	Executable string
	Host       string
	Stop       string
}

type ClientConfig struct {
	Executable string
}

type Config struct {
	CanAddHost          bool
	CanTrustCertificate string
	IsolateMetadata     bool
	IsolateProfiles     bool
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
