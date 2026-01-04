package internal

import (
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
)

var Id uuid.UUID
var AnnounceMessageData map[string]common.AnnounceMessageData002
var GeneratePlatformUserId bool

type AnnounceMessageDataLatest = common.AnnounceMessageData002
