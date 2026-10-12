package internal

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/uuid"
)

var Id uuid.UUID
var AnnounceMessageData map[string]common.AnnounceMessageData002
var GeneratePlatformUserId bool

// Sign reports whether the server signs what it serves with its own key or
// replays the signatures captured in resources/responses. Set once at startup
// from the config, before any game is created.
var Sign bool
var CanUseInternet bool
var Authentication string

type AnnounceMessageDataLatest = common.AnnounceMessageData002
