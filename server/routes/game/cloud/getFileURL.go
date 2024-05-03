package cloud

import (
	"encoding/json"
	"net/http"
	"server/files"
	i "server/internal"
)

func GetFileURL(w http.ResponseWriter, r *http.Request) {
	namesStr := r.URL.Query().Get("names")
	var names []string
	err := json.Unmarshal([]byte(namesStr), &names)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	descriptions := make(i.A, len(names))
	for j, name := range names {
		fileData := files.CloudFiles[name]
		finalPart := fileData.Key
		descriptions[j] = i.A{
			name,
			fileData.Length,
			fileData.Id,
			"https://aoe-api.worldsedgelink.com/cloudfiles/" + finalPart,
			finalPart,
		}
	}
	i.JSON(&w, i.A{0, descriptions})
}
