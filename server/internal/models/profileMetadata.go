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
	if p.gameId == common.GameAoE3 || p.gameId == common.GameAoM {
		metadata = `{"v":1,"twr":0,"wlr":0,"ai":1,"ac":0}`
	}
	return &metadata
}
