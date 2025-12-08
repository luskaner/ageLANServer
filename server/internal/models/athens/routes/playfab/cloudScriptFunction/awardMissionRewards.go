package cloudScriptFunction

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

type AwardMissionRewardsParameters struct {
	CdnPath    string
	Difficulty string
	MissionId  string
	Won        bool
}

type AwardMissionRewardsResultItemsAdded struct {
	Amount         int    `json:"amount"`
	ItemFriendlyId string `json:"itemFriendlyId"`
}

type AwardMissionRewardsResult struct {
	ItemsAdded []AwardMissionRewardsResultItemsAdded `json:"itemsAdded"`
}

type AwardMissionRewardsFunction struct {
	*playfab.CloudScriptFunctionBase[AwardMissionRewardsParameters, AwardMissionRewardsResult]
}

func (a *AwardMissionRewardsFunction) RunTyped(game models.Game, user models.User, parameters *AwardMissionRewardsParameters) *AwardMissionRewardsResult {
	// TODO: Implement only for Challenge
	return &AwardMissionRewardsResult{
		ItemsAdded: []AwardMissionRewardsResultItemsAdded{},
	}
}

func (a *AwardMissionRewardsFunction) Name() string {
	return "AwardMissionRewards"
}

func NewAwardMissionRewardsFunction() *AwardMissionRewardsFunction {
	f := &AwardMissionRewardsFunction{}
	f.CloudScriptFunctionBase = playfab.NewCloudScriptFunctionBase[AwardMissionRewardsParameters, AwardMissionRewardsResult](f)
	return f
}
