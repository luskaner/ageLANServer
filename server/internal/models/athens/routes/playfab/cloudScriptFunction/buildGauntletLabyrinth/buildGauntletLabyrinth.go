package buildGauntletLabyrinth

import (
	"time"

	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/user"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	userData "github.com/luskaner/ageLANServer/server/internal/models/playfab/data"
)

type Result struct {
	Labyrinth *user.Labyrinth
	Progress  *user.Progress
}

type Parameters struct {
	CdnPath            string
	GauntletDifficulty string
}

type Function struct {
	*playfab.CloudScriptFunctionBase[Parameters, Result]
}

func (b *Function) RunTyped(game models.Game, u models.User, parameters *Parameters) *Result {
	athensGame := game.(*athens.Game)
	athensUser := u.(*user.User)
	nodes := GenerateNumberOfNodes()
	nodeRows := GenerateNodeRows(nodes)
	blessings := RandomizedBlessings(athensGame.AllowedBlessings)
	poolIndexes := athensGame.GauntletPoolIndexByDifficulty[parameters.GauntletDifficulty]
	missionColumns := GenerateMissions(nodeRows, poolIndexes, athensGame.GauntletMissionPools, blessings)
	var missions []user.ChallengeMission
	for _, missionColumn := range missionColumns {
		for _, mission := range missionColumn {
			missions = append(missions, *mission)
		}
	}
	var id int
	data := athensUser.Data.Data()
	if labyrinth := data.Challenge.Labyrinth; labyrinth != nil {
		id = labyrinth.Value.Id + 1
	} else {
		id = 1
	}
	now := time.Now()
	data.Challenge = user.Challenge{
		Labyrinth: &userData.BaseValue[user.Labyrinth]{
			LastUpdated: userData.CustomTime{Time: now, Format: "2006-01-02T15:04:05.000Z"},
			Permission:  "Private",
			Value: user.Labyrinth{
				Id:        id,
				Dfficulty: parameters.GauntletDifficulty,
				Missions:  missions,
			},
		},
		Progress: &userData.BaseValue[user.Progress]{
			LastUpdated: userData.CustomTime{Time: now, Format: "2006-01-02T15:04:05.000Z"},
			Permission:  "Private",
			Value: user.Progress{
				Lives:             3,
				CompletedMissions: []string{},
				Inventory:         []user.ProgressInventory{},
			},
		},
	}
	data.DataVersion += 1
	go func() {
		_ = athensUser.Data.Save()
	}()
	return &Result{
		Labyrinth: &data.Challenge.Labyrinth.Value,
		Progress:  &data.Challenge.Progress.Value,
	}
}

func (b *Function) Name() string {
	return "BuildGauntletLabyrinth"
}

func NewFunction() *Function {
	f := &Function{}
	f.CloudScriptFunctionBase = playfab.NewCloudScriptFunctionBase[Parameters, Result](f)
	return f
}
