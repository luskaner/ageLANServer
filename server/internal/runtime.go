package internal

import (
	"time"

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

// CustomFiles holds the resolved /api/MatchContent/ configuration. Set once at
// startup, before any game is created, so both the model and the HTTP handlers
// see the same values without dragging the whole config around.
var CustomFilesEnabled bool
var CustomFilesFolder string
var CustomFilesMaxFileSize int64
var CustomFilesMaxTotalSize int64
var CustomFilesExpiry time.Duration

type AnnounceMessageDataLatest = common.AnnounceMessageData002
