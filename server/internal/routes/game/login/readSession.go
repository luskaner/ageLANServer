package login

import (
	"encoding/json"
	"fmt"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
	"strconv"
)

func returnData(w http.ResponseWriter, messageId uint, messages []i.A) {
	j, _ := json.Marshal(i.A{messages})
	i.RawJSON(&w, []byte(fmt.Sprintf(`%d,%s`, messageId, j)))
}

func returnError(w http.ResponseWriter) {
	returnData(w, 0, []i.A{})
}

func ReadSession(w http.ResponseWriter, r *http.Request) {
	ackId := r.FormValue("ack")
	if ackId == "" {
		returnError(w)
		return
	}
	ackIdUint, err := strconv.ParseUint(ackId, 10, 32)
	if err != nil {
		returnError(w)
		return
	}
	sess, _ := middleware.Session(r)
	messageId, messages := sess.WaitForMessages(uint(ackIdUint))
	returnData(w, messageId, messages)
}
