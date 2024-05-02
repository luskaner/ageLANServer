package extra

import (
	"aoe2DELanServer/routes/game/advertisement/extra"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
)

func ParseParameters(c *gin.Context) (*extra.Advertisement, int, []int32, []int32, []int32, []int32) {
	profileIdsStr := c.PostForm("profile_ids")
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		profileIds = []int32{}
	}
	raceIdsStr := c.PostForm("race_ids")
	var raceIds []int32
	err = json.Unmarshal([]byte(raceIdsStr), &raceIds)
	if err != nil {
		raceIds = []int32{}
	}
	statGroupIdsStr := c.PostForm("statGroup_ids")
	var statGroupIds []int32
	err = json.Unmarshal([]byte(statGroupIdsStr), &statGroupIds)
	if err != nil {
		statGroupIds = []int32{}
	}
	teamIdsStr := c.PostForm("teamIDs")
	var teamIds []int32
	err = json.Unmarshal([]byte(teamIdsStr), &teamIds)
	if err != nil {
		teamIds = []int32{}
	}
	advIdStr := c.PostForm("match_id")
	advId, err := strconv.ParseUint(advIdStr, 10, 32)
	var adv *extra.Advertisement
	if err != nil {
		advId = 0
	} else {
		adv, _ = extra.Get(uint32(advId))
	}
	length := min(len(profileIds), len(raceIds), len(statGroupIds), len(teamIds))
	return adv, length, profileIds, raceIds, statGroupIds, teamIds
}
