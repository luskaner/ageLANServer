package relationship

import (
	"net/http"
	i "server/internal"
	"server/models"
	"strconv"
)

func Ignore(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement just in memory?
	profileIdStr := r.PostFormValue("targetProfileID")

	profileId, err := strconv.ParseInt(profileIdStr, 10, 32)
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
