package extra

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/user"
	"encoding/json"
)

func GetStatGroups(idsQuery string, isProfileId bool, includeExtraProfileInfo bool) j.A {
	var ids []int32
	err := json.Unmarshal([]byte(idsQuery), &ids)
	if err != nil {
		return nil
	}

	message := j.A{0, j.A{}, j.A{}, j.A{}}

	for _, id := range ids {
		var u *user.User
		var ok bool
		if isProfileId {
			u, ok = user.GetById(id)
		} else {
			u, ok = user.GetByStatId(id)
		}
		if !ok {
			continue
		}
		message[1] = append(message[1].(j.A), j.A{
			u.GetStatId(),
			"",
			"",
			1,
			j.A{u.GetId()},
		})
		message[2] = append(message[2].(j.A), u.GetProfileInfo(false))
		if includeExtraProfileInfo {
			message[3] = append(message[3].(j.A), u.GetExtraProfileInfo())
		}
		break
	}

	return message
}
