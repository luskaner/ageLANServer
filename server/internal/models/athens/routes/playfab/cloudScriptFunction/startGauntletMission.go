package cloudScriptFunction

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

type StartGauntletMissionParameters struct {
	CdnPath   string
	MissionId string
}

type StartGauntletMissionResult struct{}

type StartGauntletMissionFunction struct {
	*playfab.CloudScriptFunctionBase[StartGauntletMissionParameters, StartGauntletMissionResult]
}

func (s *StartGauntletMissionFunction) RunTyped(game models.Game, user models.User, parameters *StartGauntletMissionParameters) *StartGauntletMissionResult {
	// TODO: Implement
	return nil
}

func (s *StartGauntletMissionFunction) Name() string {
	return "StartGauntletMission"
}

func NewStartGauntletMissionFunction() *StartGauntletMissionFunction {
	f := &StartGauntletMissionFunction{}
	f.CloudScriptFunctionBase = playfab.NewCloudScriptFunctionBase[StartGauntletMissionParameters, StartGauntletMissionResult](f)
	return f
}
