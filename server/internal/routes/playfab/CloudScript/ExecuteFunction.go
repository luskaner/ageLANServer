package CloudScript

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/server/internal/models/athens/playfab/cloudScriptFunction"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type executeFunctionRequestInitial struct {
	CustomTags              struct{}
	Entity                  *struct{}
	FunctionName            string
	GeneratePlayStreamEvent *struct{}
}

type executeFunctionRequestEnd[T any] struct {
	FunctionParameter T
}

type executeFunctionResponse struct {
	ExecutionTimeMilliseconds int64
	FunctionName              string
	FunctionResult            json.RawMessage
	FunctionResultSize        uint32
}

func ExecuteFunction(w http.ResponseWriter, r *http.Request) {
	var req executeFunctionRequestInitial
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
		var reqPar executeFunctionRequestEnd[cloudScriptFunction.AwardMissionRewardsParameters]
		err = json.Unmarshal(bodyBytes, &reqPar)
		if err != nil || req.FunctionName == "" {
			shared.RespondBadRequest(&w)
			return
		}
		t := time.Now()
		result := fn.Run(reqPar.FunctionParameter)
		duration := time.Since(t).Milliseconds()
		var resultBytes []byte
		resultBytes, err = json.Marshal(result)
		if err != nil {
			shared.RespondBadRequest(&w)
			return
		}
		shared.RespondOK(
			&w,
			&executeFunctionResponse{
				ExecutionTimeMilliseconds: duration,
				FunctionName:              req.FunctionName,
				FunctionResult:            resultBytes,
				FunctionResultSize:        uint32(len(resultBytes)),
			},
		)
	} else {
		shared.RespondBadRequest(&w)
		return
	}
}
