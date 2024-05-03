package news

import (
	"net/http"
	i "server/internal"
)

func GetNews(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
