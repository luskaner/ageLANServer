package models

import (
	"github.com/luskaner/ageLANServer/common"
)

type AvatarMetadataUpgradableDefaultData struct {
	InitialUpgradableDefaultData[*string]
	gameId string
}

func NewAvatarMetadataUpgradableDefaultData(gameId string) *AvatarMetadataUpgradableDefaultData {
	return &AvatarMetadataUpgradableDefaultData{
		InitialUpgradableDefaultData: InitialUpgradableDefaultData[*string]{},
		gameId:                       gameId,
	}
}

func (p *AvatarMetadataUpgradableDefaultData) Default() *string {
	var metadata string
	switch p.gameId {
	case common.GameAoE3, common.GameAoM:
		metadata = `{"v":1,"twr":0,"wlr":0,"ai":1,"ac":0}`
	case common.GameAoE4:
		metadata = `{"sharedHistory":1,"hardwareType":0,"inputDeviceType":0}`
	}
	return &metadata
}
