package cloudScriptFunction

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

type AwardMissionRewardsFunction struct{}

func (a *AwardMissionRewardsFunction) Run(_ AwardMissionRewardsParameters) AwardMissionRewardsResult {
	return AwardMissionRewardsResult{
		ItemsAdded: []AwardMissionRewardsResultItemsAdded{},
	}
}

func (a *AwardMissionRewardsFunction) Name() string {
	return "AwardMissionRewards"
}
