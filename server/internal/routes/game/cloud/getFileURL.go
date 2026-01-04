package cloud

import (
	"fmt"
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type getFileURLRequest struct {
	Names i.Json[[]string] `json:"names"`
}

func GetFileURL(w http.ResponseWriter, r *http.Request) {
	var req getFileURLRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{nil}})
		return
	}
	game := models.G(r)
	descriptions := make(i.A, len(req.Names.Data))
	gameTitle := game.Title()
	for j, name := range req.Names.Data {
		fileData, ok := game.Resources().CloudFiles().Value[name]
		if !ok {
			i.JSON(&w, i.A{2, i.A{nil}})
			return
		}
		finalPart := fileData.Key
		description := i.A{
			name,
			fileData.Length,
			fileData.Id,
			fmt.Sprintf("https://%s/cloudfiles/%s", r.Host, finalPart),
		}
		if gameTitle == common.GameAoE2 {
			description = append(
				description,
				finalPart,
			)
		}
		descriptions[j] = description
	}
	i.JSON(&w, i.A{0, descriptions})
}
