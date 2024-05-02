package item

import (
	i "aoe2DELanServer/internal"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetInventoryByProfileIDs(c *gin.Context) {
	profileIdsStr := c.Query("profileIDs")
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	initialData := make(i.A, len(profileIds))
	finalData := make(i.A, len(profileIds))
	finalDataArr := i.A{
		// TODO: What this mean?
		i.A{1, 0, 0, 0, 10000, 0, 0, 0, 1},
		i.A{2, 0, 1, 0, 10000, 0, 1, 1, 0},
	}
	// TODO: Understand what values mean
	for j, profileId := range profileIds {
		profileIdStr := strconv.Itoa(int(profileId))
		initialData[j] = i.A{
			profileIdStr,
			/*j.A{
				j.A{106161033, 0, 220300, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161034, 0, 220302, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161035, 0, 220304, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161036, 0, 220500, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161037, 0, 220502, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161038, 0, 210306, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161039, 0, 220301, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161040, 0, 220305, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161041, 0, 210300, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161042, 0, 220306, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161043, 0, 210301, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161044, 0, 210302, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
				j.A{106161045, 0, 210303, profileId, 1, 0, "", 1712162365, 0, -1, 0, -1},
			},*/
			i.A{},
		}
		finalData[j] = i.A{
			profileIdStr,
			finalDataArr,
		}
	}
	c.JSON(http.StatusOK, i.A{0, initialData, finalData})
}
