package cloudScriptFunction

import (
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

var Store map[string]SpecificCloudScriptFunction
var fns = []SpecificCloudScriptFunction{
	&AwardMissionRewardsFunction{},
}

type SpecificCloudScriptFunction interface {
	playfab.CloudScriptFunction[AwardMissionRewardsParameters, AwardMissionRewardsResult]
}

func init() {
	Store = make(map[string]SpecificCloudScriptFunction, len(fns))
	for _, fn := range fns {
		Store[fn.Name()] = fn
	}
}
