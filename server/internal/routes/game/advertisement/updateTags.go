package advertisement

import (
	"encoding/json"
	"fmt"
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

type tagRequest struct {
	NumericTagNames  string `schema:"numericTagNames"`
	NumericTagValues string `schema:"numericTagValues"`
	StringTagNames   string `schema:"stringTagNames"`
	StringTagValues  string `schema:"stringTagValues"`
}

type tags struct {
	NumericTagNames  []string
	NumericTagValues []int32
	StringTagNames   []string
	StringTagValues  []string
}

func parseTags(r *http.Request) (ok bool, numericTags map[string]int32, stringTags map[string]string) {
	var t tagRequest
	if err := i.Bind(r, &t); err != nil {
		return
	}
	if t.NumericTagNames == "" {
		t.NumericTagNames = "[]"
	}
	if t.NumericTagValues == "" {
		t.NumericTagValues = "[]"
	}
	if t.StringTagNames == "" {
		t.StringTagNames = "[]"
	}
	if t.StringTagValues == "" {
		t.StringTagValues = "[]"
	}
	var at tags
	jsonText := fmt.Sprintf(`{
	"NumericTagNames": %s,
	"NumericTagValues": %s,
	"StringTagNames": %s,
	"StringTagValues": %s
}`, t.NumericTagNames, t.NumericTagValues, t.StringTagNames, t.StringTagValues)
	if err := json.Unmarshal([]byte(jsonText), &at); err != nil {
		return
	}
	if (len(at.StringTagNames) != len(at.StringTagValues)) || (len(at.NumericTagNames) != len(at.NumericTagValues)) {
		return
	}
	numericTags = make(map[string]int32, len(at.NumericTagNames))
	stringTags = make(map[string]string, len(at.StringTagNames))
	for j := 0; j < len(at.NumericTagNames); j++ {
		numericTags[at.NumericTagNames[j]] = at.NumericTagValues[j]
	}
	for j := 0; j < len(at.StringTagNames); j++ {
		stringTags[at.StringTagNames[j]] = at.StringTagValues[j]
	}
	ok = true
	return
}

func updateTagsReturnError(w *http.ResponseWriter) {
	i.JSON(w, i.A{2})
}

func UpdateTags(w http.ResponseWriter, r *http.Request) {
	ok, numericTags, stringTags := parseTags(r)
	if !ok {
		updateTagsReturnError(&w)
		return
	}
	var q shared.AdvertisementId
	if err := i.Bind(r, &q); err != nil {
		updateTagsReturnError(&w)
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	matchingAdv, foundAdv := advertisements.GetAdvertisement(q.AdvertisementId)
	if !foundAdv {
		updateTagsReturnError(&w)
		return
	}
	sess := middleware.SessionOrPanic(r)
	advertisements.WithWriteLock(matchingAdv.GetId(), func() {
		if matchingAdv.GetHostId() != sess.GetUserId() {
			updateTagsReturnError(&w)
			ok = false
			return
		}
		matchingAdv.UnsafeUpdateTags(numericTags, stringTags)
	})
	if !ok {
		return
	}
	i.JSON(&w, i.A{0})
}
