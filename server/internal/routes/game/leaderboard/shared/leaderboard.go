package shared

import (
	"encoding/json"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetStatGroups(r *http.Request, idsQuery string, isProfileId bool, includeExtraProfileInfo bool) i.A {
	var ids []int32
	err := json.Unmarshal([]byte(idsQuery), &ids)
	if err != nil {
		return nil
	}

	message := i.A{0, i.A{}, i.A{}, i.A{}}
	users := models.G(r).Users()
	for _, id := range ids {
		var u *models.MainUser
		var ok bool
		if isProfileId {
			u, ok = users.GetUserById(id)
		} else {
			u, ok = users.GetUserByStatId(id)
		}
		if !ok {
			continue
		}
		message[1] = append(message[1].(i.A), i.A{
			u.GetStatId(),
			"",
			"",
			1,
			i.A{u.GetId()},
		})
		message[2] = append(message[2].(i.A), u.GetProfileInfo(false))
		if includeExtraProfileInfo {
			message[3] = append(message[3].(i.A), u.GetExtraProfileInfo())
		}
		break
	}

	return message
}
