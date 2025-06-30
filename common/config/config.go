package config

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/viper"
	"strings"
)

func InitConfig(v *viper.Viper, paths []string, files []string, envPrefixSuffix string, printerFn func(path string)) {
	if printerFn == nil {
		printerFn = func(_ string) {}
	}
	for _, configPath := range paths {
		v.AddConfigPath(configPath)
	}
	v.SetConfigType("toml")
	for _, cfgFile := range files {
		v.SetConfigFile(cfgFile)
		if err := v.MergeInConfig(); err == nil {
			printerFn(v.ConfigFileUsed())
		}
	}
	if v.ConfigFileUsed() == "" {
		v.SetConfigName("config")
		if err := viper.ReadInConfig(); err == nil {
			printerFn(v.ConfigFileUsed())
		}
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix(strings.ToUpper(common.Name + "_" + envPrefixSuffix))
	v.AutomaticEnv()
}
