package relationship

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strconv"
)

func Ignore(w http.ResponseWriter, r *http.Request) {
	profileIdStr := r.PostFormValue("targetProfileID")

	profileId, err := strconv.ParseInt(profileIdStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	u, ok := models.G(r).Users().GetUserById(int32(profileId))
	if !ok {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	i.JSON(&w, i.A{2, u.GetProfileInfo(false), i.A{}})
}
