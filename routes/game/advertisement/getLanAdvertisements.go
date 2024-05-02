package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"aoe2DELanServer/session"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type modDll struct {
	File     string `form:"modDLLFile"`
	Checksum uint32 `form:"modDLLChecksum"`
}

type query struct {
	AppBinaryChecksum uint32 `form:"appBinaryChecksum"`
	DataChecksum      uint32 `form:"dataChecksum"`
	MatchType         uint8  `form:"matchType_id"`
	ModDll            modDll
	ModName           string `form:"modName"`
	ModVersion        string `form:"modVersion"`
	VersionFlags      uint32 `form:"versionFlags"`
	RelayRegions      string `form:"lanServerGuids"`
}

func GetLanAdvertisements(c *gin.Context) {
	var q query
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}, j.A{}})
		return
	}
	var lanServerGuids []string
	err := json.Unmarshal([]byte(q.RelayRegions), &lanServerGuids)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}, j.A{}})
		return
	}
	sessAny, _ := c.Get("session")
	sess, ok := sessAny.(*session.Info)
	if !ok {
		c.JSON(http.StatusOK, j.A{2, j.A{}, j.A{}})
		return
	}
	lanServerGuidsMap := make(map[string]struct{}, len(lanServerGuids))
	for _, guid := range lanServerGuids {
		lanServerGuidsMap[guid] = struct{}{}
	}
	currentUser := sess.GetUser()
	advs := extra.FindAdvertisementsEncoded(func(adv *extra.Advertisement) bool {
		_, relayRegionMatches := lanServerGuidsMap[adv.GetRelayRegion()]
		_, isPeer := adv.GetPeer(currentUser)
		return adv.GetJoinable() &&
			adv.GetVisible() &&
			!isPeer &&
			adv.GetAppBinaryChecksum() == q.AppBinaryChecksum &&
			adv.GetDataChecksum() == q.DataChecksum &&
			adv.GetMatchType() == q.MatchType &&
			adv.GetModDllFile() == q.ModDll.File &&
			adv.GetModDllChecksum() == q.ModDll.Checksum &&
			adv.GetModName() == q.ModName &&
			adv.GetModVersion() == q.ModVersion &&
			adv.GetVersionFlags() == q.VersionFlags &&
			relayRegionMatches
	})
	if advs == nil {
		c.JSON(http.StatusOK,
			j.A{0, j.A{}, j.A{}},
		)
	} else {
		c.JSON(http.StatusOK,
			j.A{0, advs, j.A{}},
		)
	}
}
