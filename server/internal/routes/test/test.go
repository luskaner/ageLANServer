package test

import (
	"net/http"
	"strconv"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func Test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.IdHeader, i.Id.String())
	w.Header().Set(common.VersionHeader, strconv.Itoa(common.AnnounceVersionLatest))
	i.JSON(&w, i.AnnounceMessageData[models.G(r).Title()])
}
