package routes

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models/apiAgeOfEmpires"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	apiAgeOfEmpires2 "github.com/luskaner/ageLANServer/server/internal/routes/apiAgeOfEmpires"
	"github.com/luskaner/ageLANServer/server/internal/routes/apiAgeOfEmpires/textmoderation"
	cacertPem "github.com/luskaner/ageLANServer/server/internal/routes/cacert.pem"
	"github.com/luskaner/ageLANServer/server/internal/routes/cloudfiles"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/account"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/achievement"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement"
	Automatch2 "github.com/luskaner/ageLANServer/server/internal/routes/game/automatch2"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/challenge"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/chat"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/clan"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/cloud"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/communityEvent"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/invitation"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/item"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/leaderboard"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/leaderboard/age3"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/login"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/news"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/party"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/playerreport"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/relationship"
	"github.com/luskaner/ageLANServer/server/internal/routes/msstore"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Event"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Inventory"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/MultiplayerServer"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Party"
	"github.com/luskaner/ageLANServer/server/internal/routes/test"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type Group struct {
	parent *Group
	path   string
	mux    *http.ServeMux
}

func (g *Group) fullPath() string {
	if g.parent == nil {
		return g.path
	}
	return g.parent.fullPath() + g.path
}

func (g *Group) Subgroup(path string) *Group {
	return &Group{
		parent: g,
		path:   path,
		mux:    g.mux,
	}
}

func (g *Group) HandleFunc(method string, path string, handler http.HandlerFunc) {
	g.mux.HandleFunc(method+" "+g.fullPath()+path, handler)
}

func (g *Group) Handle(method string, path string, handler http.Handler) {
	g.mux.Handle(method+" "+g.fullPath()+path, handler)
}

func (g *Group) HandlePath(path string, handler http.Handler) {
	g.mux.Handle(g.fullPath()+path, handler)
}

func Initialize(mux *http.ServeMux, game string) {
	baseGroup := Group{
		path: "",
		mux:  mux,
	}
	gameGroup := baseGroup.Subgroup("/game")
	itemGroup := gameGroup.Subgroup("/item")
	itemGroup.HandleFunc("GET", "/getItemBundleItemsJson", item.GetItemBundleItemsJson)
	itemGroup.HandleFunc("GET", "/getItemDefinitionsJson", item.GetItemDefinitionsJson)
	itemGroup.HandleFunc("GET", "/getItemLoadouts", item.GetItemLoadouts)
	itemGroup.HandleFunc("POST", "/signItems", item.SignItems)
	itemGroup.HandleFunc("GET", "/getInventoryByProfileIDs", item.GetInventoryByProfileIDs)
	itemGroup.HandleFunc("POST", "/detachItems", item.DetachItems)

	clanGroup := gameGroup.Subgroup("/clan")
	clanGroup.HandleFunc("POST", "/create", clan.Create)
	clanGroup.HandleFunc("GET", "/find", clan.Find)

	communityEventGroup := gameGroup.Subgroup("/CommunityEvent")
	communityEventGroup.HandleFunc("GET", "/getAvailableCommunityEvents", communityEvent.GetAvailableCommunityEvents)

	challengeGroup := gameGroup.Subgroup("/challenge")
	if game == common.GameAoE3 {
		challengeGroup.HandleFunc("POST", "/updateProgress", challenge.UpdateProgress)
	}
	if game == common.GameAoM {
		challengeGroup.HandleFunc("POST", "/updateProgressBatched", challenge.UpdateProgressBatched)
	}
	ChallengeGroup := gameGroup.Subgroup("/Challenge")
	if game == common.GameAoE3 {
		ChallengeGroup.HandleFunc("POST", "/getChallengeProgress", challenge.GetChallengeProgress)
	}
	if game == common.GameAoE2 || game == common.GameAoM {
		ChallengeGroup.HandleFunc("GET", "/getChallengeProgress", challenge.GetChallengeProgress)
	}
	ChallengeGroup.HandleFunc("GET", "/getChallenges", challenge.GetChallenges)

	newsGroup := gameGroup.Subgroup("/news")
	newsGroup.HandleFunc("GET", "/getNews", news.GetNews)

	loginGroup := gameGroup.Subgroup("/login")
	loginGroup.HandleFunc("POST", "/platformlogin", login.Platformlogin)
	loginGroup.HandleFunc("POST", "/logout", login.Logout)
	loginGroup.HandleFunc("POST", "/readSession", login.ReadSession)
	accountGroup := gameGroup.Subgroup("/account")
	accountGroup.HandleFunc("POST", "/setLanguage", account.SetLanguage)
	accountGroup.HandleFunc("POST", "/setCrossplayEnabled", account.SetCrossplayEnabled)
	accountGroup.HandleFunc("POST", "/setAvatarMetadata", account.SetAvatarMetadata)
	accountGroup.HandleFunc("POST", "/FindProfilesByPlatformID", account.FindProfilesByPlatformID)
	accountGroup.HandleFunc("GET", "/FindProfiles", account.FindProfiles)
	accountGroup.HandleFunc("GET", "/getProfileName", account.GetProfileName)
	if game == common.GameAoE3 || game == common.GameAoM {
		accountGroup.HandleFunc("GET", "/getProfileProperty", account.GetProfileProperty)
	}

	LeaderboardGroup := gameGroup.Subgroup("/Leaderboard")
	if game == common.GameAoE3 {
		LeaderboardGroup.HandleFunc("POST", "/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
	}
	if game == common.GameAoE2 || game == common.GameAoM {
		LeaderboardGroup.HandleFunc("GET", "/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
	}
	LeaderboardGroup.HandleFunc("GET", "/getLeaderBoard", leaderboard.GetLeaderBoard)
	LeaderboardGroup.HandleFunc("GET", "/getAvailableLeaderboards", leaderboard.GetAvailableLeaderboards)
	LeaderboardGroup.HandleFunc("GET", "/getStatGroupsByProfileIDs", leaderboard.GetStatGroupsByProfileIDs)
	LeaderboardGroup.HandleFunc("GET", "/getStatsForLeaderboardByProfileName", leaderboard.GetStatsForLeaderboardByProfileName)
	LeaderboardGroup.HandleFunc("GET", "/getPartyStat", leaderboard.GetPartyStat)
	if game == common.GameAoE3 {
		LeaderboardGroup.HandleFunc("GET", "/getAvatarStatLeaderBoard", leaderboard.GetAvatarStatLeaderBoard)
	}
	leaderboardGroup := gameGroup.Subgroup("/leaderboard")
	leaderboardGroup.HandleFunc("POST", "/applyOfflineUpdates", leaderboard.ApplyOfflineUpdates)
	var setStatValues func(http.ResponseWriter, *http.Request)
	if game == common.GameAoE3 {
		setStatValues = age3.SetAvatarStatValues
	} else {
		setStatValues = leaderboard.SetAvatarStatValues
	}
	leaderboardGroup.HandleFunc("POST", "/setAvatarStatValues", setStatValues)

	automatch2Group := gameGroup.Subgroup("/automatch2")
	automatch2Group.HandleFunc("GET", "/getAutomatchMap", Automatch2.GetAutomatchMap)

	AchievementGroup := gameGroup.Subgroup("/Achievement")
	AchievementGroup.HandleFunc("GET", "/getAchievements", achievement.GetAchievements)
	AchievementGroup.HandleFunc("GET", "/getAvailableAchievements", achievement.GetAvailableAchievements)

	achievementGroup := gameGroup.Subgroup("/achievement")
	achievementGroup.HandleFunc("POST", "/applyOfflineUpdates", achievement.ApplyOfflineUpdates)
	achievementGroup.HandleFunc("POST", "/grantAchievement", achievement.GrantAchievement)
	achievementGroup.HandleFunc("POST", "/syncStats", achievement.SyncStats)

	advertisementGroup := gameGroup.Subgroup("/advertisement")
	if game == common.GameAoE2 || game == common.GameAoM {
		advertisementGroup.HandleFunc("POST", "/updatePlatformSessionID", advertisement.UpdatePlatformSessionID)
	}
	advertisementGroup.HandleFunc("POST", "/join", advertisement.Join)
	if game == common.GameAoE2 || game == common.GameAoM {
		advertisementGroup.HandleFunc("POST", "/updateTags", advertisement.UpdateTags)
	}
	advertisementGroup.HandleFunc("POST", "/update", advertisement.Update)
	advertisementGroup.HandleFunc("POST", "/leave", advertisement.Leave)
	advertisementGroup.HandleFunc("POST", "/host", advertisement.Host)
	if game == common.GameAoE1 || game == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/getLanAdvertisements", advertisement.GetLanAdvertisements)
	}
	if game == common.GameAoE2 {
		advertisementGroup.HandleFunc("GET", "/getLanAdvertisements", advertisement.GetLanAdvertisements)
	}
	if game == common.GameAoE1 || game == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/updatePlatformLobbyID", advertisement.UpdatePlatformLobbyID)
	}
	if game == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
	}
	if game == common.GameAoE2 || game == common.GameAoM {
		advertisementGroup.HandleFunc("GET", "/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
	}
	advertisementGroup.HandleFunc("GET", "/getAdvertisements", advertisement.GetAdvertisements)
	if game == common.GameAoE1 || game == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/findAdvertisements", advertisement.FindAdvertisements)
	}
	if game == common.GameAoE2 || game == common.GameAoM {
		advertisementGroup.HandleFunc("GET", "/findAdvertisements", advertisement.FindAdvertisements)
	}

	advertisementGroup.HandleFunc("POST", "/updateState", advertisement.UpdateState)

	if game == common.GameAoE2 || game == common.GameAoE3 || game == common.GameAoM {
		advertisementGroup.HandleFunc("POST", "/startObserving", advertisement.StartObserving)
		advertisementGroup.HandleFunc("POST", "/stopObserving", advertisement.StopObserving)
	}

	chatGroup := gameGroup.Subgroup("/chat")
	if game == common.GameAoE1 || game == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/getChatChannels", chat.GetChatChannels)
	}
	if game == common.GameAoE2 || game == common.GameAoM {
		chatGroup.HandleFunc("GET", "/getChatChannels", chat.GetChatChannels)
	}
	chatGroup.HandleFunc("GET", "/getOfflineMessages", chat.GetOfflineMessages)
	if game == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/joinChannel", chat.JoinChannel)
	}
	if game == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/leaveChannel", chat.LeaveChannel)
	}
	if game == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/sendText", chat.SendText)
	}
	if game == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/sendWhisper", chat.SendWhisper)
	}
	if game == common.GameAoM {
		chatGroup.HandleFunc("POST", "/sendWhispers", chat.SendWhisper)
	}
	if game == common.GameAoM {
		chatGroup.HandleFunc("POST", "/deleteOfflineMessage", chat.DeleteOfflineMessage)
	}

	relationshipGroup := gameGroup.Subgroup("/relationship")
	if game == common.GameAoE1 || game == common.GameAoE3 {
		relationshipGroup.HandleFunc("POST", "/getRelationships", relationship.GetRelationships)
	}
	if game == common.GameAoE2 || game == common.GameAoM {
		relationshipGroup.HandleFunc("GET", "/getRelationships", relationship.GetRelationships)
	}
	relationshipGroup.HandleFunc("GET", "/getPresenceData", relationship.GetPresenceData)
	relationshipGroup.HandleFunc("POST", "/setPresence", relationship.SetPresence)
	if game == common.GameAoE3 || game == common.GameAoM {
		relationshipGroup.HandleFunc("POST", "/setPresenceProperty", relationship.SetPresenceProperty)
	}
	if game == common.GameAoE3 || game == common.GameAoM {
		relationshipGroup.HandleFunc("POST", "/addfriend", relationship.Addfriend)
	}
	relationshipGroup.HandleFunc("POST", "/ignore", relationship.Ignore)
	relationshipGroup.HandleFunc("POST", "/clearRelationship", relationship.ClearRelationship)

	partyGroup := gameGroup.Subgroup("/party")
	partyGroup.HandleFunc("POST", "/peerAdd", party.PeerAdd)
	partyGroup.HandleFunc("POST", "/peerUpdate", party.PeerUpdate)
	partyGroup.HandleFunc("POST", "/sendMatchChat", party.SendMatchChat)
	partyGroup.HandleFunc("POST", "/reportMatch", party.ReportMatch)
	partyGroup.HandleFunc("POST", "/finalizeReplayUpload", party.FinalizeReplayUpload)
	partyGroup.HandleFunc("POST", "/updateHost", party.UpdateHost)
	if game == common.GameAoM {
		partyGroup.HandleFunc("POST", "/createOrReportSinglePlayer", party.CreateOrReportSinglePlayer)
	}
	playerReportGroup := gameGroup.Subgroup("/playerreport")
	// TODO: Check if it applies to AoE I/AoE III
	if game == common.GameAoE2 || game == common.GameAoM {
		playerReportGroup.HandleFunc("POST", "/reportUser", playerreport.ReportUser)
	}

	invitationGroup := gameGroup.Subgroup("/invitation")
	invitationGroup.HandleFunc("POST", "/extendInvitation", invitation.ExtendInvitation)
	invitationGroup.HandleFunc("POST", "/cancelInvitation", invitation.CancelInvitation)
	invitationGroup.HandleFunc("POST", "/replyToInvitation", invitation.ReplyToInvitation)

	cloudGroup := gameGroup.Subgroup("/cloud")
	if game == common.GameAoE3 {
		cloudGroup.HandleFunc("POST", "/getFileURL", cloud.GetFileURL)
	}
	// TODO: Enable to AoM if/when it gets cloud support
	if game == common.GameAoE2 {
		cloudGroup.HandleFunc("GET", "/getFileURL", cloud.GetFileURL)
	}

	cloudGroup.HandleFunc("GET", "/getTempCredentials", cloud.GetTempCredentials)

	msstoreGroup := gameGroup.Subgroup("/msstore")
	msstoreGroup.HandleFunc("GET", "/getStoreTokens", msstore.GetStoreTokens)

	// Used for the launcher
	baseGroup.HandleFunc("GET", "/test", test.Test)
	baseGroup.HandleFunc("GET", "/cacert.pem", cacertPem.CacertPem)

	baseGroup.HandleFunc("GET", "/wss/", wss.Handle)
	// TODO: Enable to AoM if/when it gets cloud support
	if game == common.GameAoE2 || game == common.GameAoE3 {
		baseGroup.HandleFunc("GET", "/cloudfiles/", cloudfiles.Cloudfiles)
	}

	if game == common.GameAoM {
		playfabGroup := baseGroup.Subgroup(playfab.Prefix)

		playfabClientGroup := playfabGroup.Subgroup("/Client")
		playfabClientGroup.HandleFunc("POST", "/GetPlayerCombinedInfo", Client.GetPlayerCombinedInfo)
		playfabClientGroup.HandleFunc("POST", "/GetTime", Client.GetTime)
		playfabClientGroup.HandleFunc("POST", "/GetTitleData", Client.GetTitleData)
		playfabClientGroup.HandleFunc("POST", "/GetUserReadOnlyData", Client.GetUserReadOnlyData)
		playfabClientGroup.HandleFunc("POST", "/LoginWithSteam", Client.LoginWithSteam)
		playfabClientGroup.HandleFunc("POST", "/UpdateUserTitleDisplayName", Client.UpdateUserTitleDisplayName)

		playfabEventGroup := playfabGroup.Subgroup("/Event")
		playfabEventGroup.HandleFunc("POST", "/WriteTelemetryEvents", Event.WriteTelemetryEvents)

		playfabInventoryGroup := playfabGroup.Subgroup("/Inventory")
		playfabInventoryGroup.HandleFunc("POST", "/GetInventoryItems", Inventory.GetInventoryItems)

		playfabMultiplayerServerGroup := playfabGroup.Subgroup("/MultiplayerServer")
		playfabMultiplayerServerGroup.HandleFunc("POST", "/GetCognitiveServicesToken", MultiplayerServer.GetCognitiveServicesToken)
		playfabMultiplayerServerGroup.HandleFunc("POST", "/ListPartyQosServers", MultiplayerServer.ListPartyQosServers)

		playfabPartyGroup := playfabGroup.Subgroup("/Party")
		playfabPartyGroup.HandleFunc("POST", "/RequestParty", Party.RequestParty)
		fs := http.FileServer(http.Dir(playfab.BaseDir))
		playfabGroup.Handle(
			"GET",
			playfab.StaticSuffix+"/",
			http.StripPrefix(playfab.StaticPrefix+"/", fs),
		)
	}

	if game == common.GameAoE3 || game == common.GameAoM {
		apiAgeOfEmpiresGroup := baseGroup.Subgroup(apiAgeOfEmpires.Prefix)
		apiAgeOfEmpiresGroup.HandleFunc("POST", "/textmoderation", textmoderation.TextModeration)
		if proxy := apiAgeOfEmpires2.Root(); proxy != nil {
			apiAgeOfEmpiresGroup.HandlePath("/", proxy)
		}
	}
}
