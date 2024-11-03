package age3

import (
	"encoding/json"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"time"
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
	response := make([]i.A, min(len(avatarStatIds), len(values)))
	users := models.G(r).Users()

	for j := 0; j < len(response); j++ {
		u, _ := users.GetUserByStatId(avatarStatIds[j])
		response[j] = i.A{avatarStatIds[j], u.GetId(), values[j], "", time.Now().UTC().Unix()}
	}
	i.JSON(&w, i.A{0, i.A{0}, response})
}
