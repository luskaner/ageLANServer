package relationship

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"net/http"
	"strconv"
)

func Ignore(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement just in memory?
	profileIdStr := r.PostFormValue("targetProfileID")
	profileId, err := strconv.Atoi(profileIdStr)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	u, ok := models.GetUserById(int32(profileId))
	if !ok {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	i.JSON(&w, i.A{2, u.GetProfileInfo(false), i.A{}})
}
