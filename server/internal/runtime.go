package internal

import (
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/viper"
)

var Version string
var Id uuid.UUID

func AnnounceMessageData() common.AnnounceMessageData002 {
	return AnnounceMessageDataLatest{
		AnnounceMessageData001: common.AnnounceMessageData001{
			GameIds: viper.GetStringSlice("Games"),
		},
		Version: Version,
	}
}
