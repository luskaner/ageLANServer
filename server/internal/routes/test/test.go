package test

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
	"strconv"
)

func Test(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(common.IdHeader, i.Id.String())
	w.Header().Set(common.VersionHeader, strconv.Itoa(i.AnnounceVersionLatest))
	i.JSON(&w, i.AnnounceMessageData)
}
