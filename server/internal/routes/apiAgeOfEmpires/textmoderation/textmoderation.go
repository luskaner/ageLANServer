package textmoderation

import (
	"encoding/json"
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

type textModerationRequest struct {
	ConversationId string `json:"conversationId"`
	TextContent    string `json:"textContent"`
	Language       string `json:"language"`
	TextType       string `json:"textType"`
}

type textModerationResponse struct {
	FilterResult         string `json:"filterResult"`
	FamilyFriendlyResult string `json:"familyFriendlyResult"`
	MediumResult         string `json:"mediumResult"`
	MatureResult         string `json:"matureResult"`
	MaturePlusResult     string `json:"maturePlusResult"`
	TranslationAvailable bool   `json:"translationAvailable"`
}

func TextModeration(w http.ResponseWriter, r *http.Request) {
	var req textModerationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
	if req.TextType == "SanitisationUsername" {
		i.JSON(
			&w,
			textModerationResponse{
				FilterResult:         "Allow",
				FamilyFriendlyResult: "Allow",
				MediumResult:         "Allow",
				MatureResult:         "Allow",
				MaturePlusResult:     "Allow",
				TranslationAvailable: false,
			},
		)
	}
}
