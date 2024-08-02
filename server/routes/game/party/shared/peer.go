package shared

import (
	"encoding/json"
	"github.com/luskaner/aoe2DELanServer/server/models"
	"net/http"
	"strconv"
)

func ParseParameters(r *http.Request) (*models.Advertisement, int, []int32, []int32, []int32, []int32) {
	profileIdsStr := r.PostFormValue("profile_ids")
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		profileIds = []int32{}
	}
	raceIdsStr := r.PostFormValue("race_ids")
	var raceIds []int32
	err = json.Unmarshal([]byte(raceIdsStr), &raceIds)
	if err != nil {
		raceIds = []int32{}
	}
	statGroupIdsStr := r.PostFormValue("statGroup_ids")
	var statGroupIds []int32
	err = json.Unmarshal([]byte(statGroupIdsStr), &statGroupIds)
	if err != nil {
		statGroupIds = []int32{}
	}
	teamIdsStr := r.PostFormValue("teamIDs")
	var teamIds []int32
	err = json.Unmarshal([]byte(teamIdsStr), &teamIds)
	if err != nil {
		teamIds = []int32{}
	}
	advIdStr := r.PostFormValue("match_id")
	advId, err := strconv.ParseInt(advIdStr, 10, 32)
	var adv *models.Advertisement
	if err == nil {
		adv, _ = models.GetAdvertisement(int32(advId))
	}
	length := min(len(profileIds), len(raceIds), len(statGroupIds), len(teamIds))
	return adv, length, profileIds, raceIds, statGroupIds, teamIds
}
