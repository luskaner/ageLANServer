package cmdUtils

import (
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
)

func ExistingServers(gameId string) (err error, names mapset.Set[string], regions mapset.Set[string]) {
	var configs []battleServerConfig.Config
	configs, err = battleServerConfig.Configs(gameId, true)
	if err != nil {
		return
	}
	names = mapset.NewThreadUnsafeSetWithSize[string](len(configs))
	regions = mapset.NewThreadUnsafeSetWithSize[string](len(configs))
	for _, config := range configs {
		names.Add(strings.ToLower(config.Name))
		regions.Add(strings.ToLower(config.Region))
	}
	return
}
