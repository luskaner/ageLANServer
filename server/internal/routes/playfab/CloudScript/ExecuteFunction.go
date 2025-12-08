package CloudScript

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/routes/playfab/cloudScriptFunction"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type executeFunctionRequest struct {
	CustomTags              struct{}
	Entity                  *struct{}
	FunctionName            string
	GeneratePlayStreamEvent *struct{}
	FunctionParameter       json.RawMessage
}

type executeFunctionResponse struct {
	ExecutionTimeMilliseconds int64
	FunctionName              string
	FunctionResult            json.RawMessage
	FunctionResultSize        uint32
}

func ExecuteFunction(w http.ResponseWriter, r *http.Request) {
	var req executeFunctionRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		shared.RespondBadRequest(&w)
		return
	}
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil || req.FunctionName == "" {
		shared.RespondBadRequest(&w)
		return
	}
	if fn, ok := cloudScriptFunction.Store[req.FunctionName]; ok {
		reqPar := fn.NewParameters()
		err = json.Unmarshal(req.FunctionParameter, &reqPar)
		if err != nil {
			shared.RespondBadRequest(&w)
			return
		}
		u := playfab.SessionOrPanic(r).User()
		game := models.G(r)
		t := time.Now()
		result := fn.Run(game, u, reqPar)
		duration := time.Since(t).Milliseconds()
		var resultSize uint32
		var resultBytes []byte
		if result != nil {
			resultBytes, err = json.Marshal(result)
			if err != nil {
				shared.RespondBadRequest(&w)
				return
			}
			resultSize = uint32(len(resultBytes))
		}
		shared.RespondOK(
			&w,
			&executeFunctionResponse{
				ExecutionTimeMilliseconds: duration,
				FunctionName:              req.FunctionName,
				FunctionResult:            resultBytes,
				FunctionResultSize:        resultSize,
			},
		)
	} else {
		shared.RespondBadRequest(&w)
		return
	}
}
