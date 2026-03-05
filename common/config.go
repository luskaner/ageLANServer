package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// KoanfFileLoadError wraps parser/load errors for a specific config file path.
type KoanfFileLoadError struct {
	Path string
	Err  error
}

func (e *KoanfFileLoadError) Error() string {
	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

func (e *KoanfFileLoadError) Unwrap() error {
	return e.Err
}

// LoadKoanfLayers applies config layers in this order:
// defaults < env < file < flags.
func LoadKoanfLayers(
	k *koanf.Koanf,
	defaults map[string]any,
	fileCandidates []string,
	parser koanf.Parser,
	fs *pflag.FlagSet,
	fsBindings map[string]string,
	envPrefix string,
) (string, error) {
	if k == nil {
		return "", fmt.Errorf("koanf instance is nil")
	}

	_ = k.Load(confmap.Provider(defaults, "."), nil)
	_ = k.Load(koanfEnvProvider(Name+envPrefix), nil)

	usedFile := ""
	for _, candidate := range fileCandidates {
		if candidate == "" {
			continue
		}
		if err := k.Load(file.Provider(candidate), parser); err == nil {
			usedFile = candidate
			break
		} else if !os.IsNotExist(err) {
			return "", &KoanfFileLoadError{Path: candidate, Err: err}
		}
	}

	var posFlag *posflag.Posflag
	if fsBindings == nil {
		posFlag = posflag.Provider(fs, ".", k)
	} else {
		posFlag = posflag.ProviderWithFlag(
			fs,
			".",
			k,
			func(f *pflag.Flag) (string, interface{}) {
				key := f.Name
				if binding, ok := fsBindings[key]; ok {
					key = binding
				}
				return key, posflag.FlagVal(fs, f)
			})
	}
	_ = k.Load(posFlag, nil)
	return usedFile, nil
}

// koanfEnvProvider returns a provider that maps ENV keys to koanf keys using '.' delimiters.
// It lowercases keys and replaces '_' with '.'. Values with spaces become string slices.
func koanfEnvProvider(prefix string) *env.Env {
	finalPrefix := strings.ReplaceAll(strings.ToUpper(prefix), "-", "_") + "_"
	return env.Provider(".", env.Opt{
		Prefix: finalPrefix,
		TransformFunc: func(key, value string) (string, any) {
			k := strings.TrimPrefix(key, finalPrefix)
			k = strings.ReplaceAll(k, "_", ".")
			if strings.Contains(value, " ") {
				return k, strings.Split(value, " ")
			}
			return k, value
		},
	})
}
