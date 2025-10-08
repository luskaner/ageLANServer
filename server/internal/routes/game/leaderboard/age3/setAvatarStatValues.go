package age3

import (
	"encoding/json"
	"net/http"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func SetAvatarStatValues(w http.ResponseWriter, r *http.Request) {
	avatarStatIdsStr := r.URL.Query().Get("avatarStat_ids")
	if len(avatarStatIdsStr) < 1 {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	var avatarStatIds []int32
	err := json.Unmarshal([]byte(avatarStatIdsStr), &avatarStatIds)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	valuesStr := r.URL.Query().Get("values")
	if len(avatarStatIdsStr) < 1 {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	var values []int32
	err = json.Unmarshal([]byte(valuesStr), &values)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	if len(avatarStatIds) != len(values) {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	response := make([]i.A, len(avatarStatIds))
	users := models.G(r).Users()

	for j := 0; j < len(response); j++ {
		u, ok := users.GetUserByStatId(avatarStatIds[j])
		if !ok {
			i.JSON(&w, i.A{2, i.A{}, i.A{}})
			return
		}
		response[j] = i.A{avatarStatIds[j], u.GetId(), values[j], "", time.Now().UTC().Unix()}
	}
	i.JSON(&w, i.A{0, i.A{0}, response})
}
