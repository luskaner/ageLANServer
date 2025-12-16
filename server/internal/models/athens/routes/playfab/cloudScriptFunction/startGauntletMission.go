package cloudScriptFunction

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/user"
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

func (s *StartGauntletMissionFunction) RunTyped(_ models.Game, u models.User, parameters *StartGauntletMissionParameters) *StartGauntletMissionResult {
	athensUser := u.(*user.User)
	d := athensUser.Data
	finalData := d.Data()
	progress := finalData.Challenge.Progress
	if progress == nil {
		return nil
	}
	if (*progress.Value).MissionBeingPlayedRightNow != parameters.MissionId {
		progress.Update(func(progress *user.Progress) {
			progress.MissionBeingPlayedRightNow = parameters.MissionId
		})
		finalData.DataVersion++
		defer func() {
			_ = d.Save()
		}()
	}
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
