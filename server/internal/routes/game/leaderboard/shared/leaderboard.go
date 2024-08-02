package shared

import (
	"encoding/json"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
)

func GetStatGroups(idsQuery string, isProfileId bool, includeExtraProfileInfo bool) i.A {
	var ids []int32
	err := json.Unmarshal([]byte(idsQuery), &ids)
	if err != nil {
		return nil
	}

	message := i.A{0, i.A{}, i.A{}, i.A{}}

	for _, id := range ids {
		var u *models.User
		var ok bool
		if isProfileId {
			u, ok = models.GetUserById(id)
		} else {
			u, ok = models.GetUserByStatId(id)
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
