package router

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
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
	"github.com/luskaner/ageLANServer/server/internal/routes/game/login"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/news"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/party"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/playerreport"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/relationship"
	"github.com/luskaner/ageLANServer/server/internal/routes/msstore"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type Game struct {
	Router
}

func (g *Game) Name() string {
	return "game"
}

func (g *Game) InitializeRoutes(gameId string, _ http.Handler) http.Handler {
	g.initialize()
	gameGroup := g.group.Subgroup("/game")
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

	if gameId == common.GameAoM {
		communityEventGroup.HandleFunc("GET", "/getEventStats", communityEvent.GetEventStats)
		communityEventGroup.HandleFunc("GET", "/getEventLeaderboard", communityEvent.GetEventLeaderboard)
	}

	challengeGroup := gameGroup.Subgroup("/challenge")
	if gameId == common.GameAoE3 {
		challengeGroup.HandleFunc("POST", "/updateProgress", challenge.UpdateProgress)
	}
	if gameId == common.GameAoM {
		challengeGroup.HandleFunc("POST", "/updateProgressBatched", challenge.UpdateProgressBatched)
	}
	ChallengeGroup := gameGroup.Subgroup("/Challenge")
	if gameId == common.GameAoE3 {
		ChallengeGroup.HandleFunc("POST", "/getChallengeProgress", challenge.GetChallengeProgress)
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
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
	if gameId == common.GameAoE3 || gameId == common.GameAoM {
		accountGroup.HandleFunc("GET", "/getProfileProperty", account.GetProfileProperty)
	}

	LeaderboardGroup := gameGroup.Subgroup("/Leaderboard")
	if gameId == common.GameAoE3 {
		LeaderboardGroup.HandleFunc("POST", "/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		LeaderboardGroup.HandleFunc("GET", "/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
	}
	LeaderboardGroup.HandleFunc("GET", "/getLeaderBoard", leaderboard.GetLeaderBoard)
	LeaderboardGroup.HandleFunc("GET", "/getAvailableLeaderboards", leaderboard.GetAvailableLeaderboards)
	LeaderboardGroup.HandleFunc("GET", "/getStatGroupsByProfileIDs", leaderboard.GetStatGroupsByProfileIDs)
	LeaderboardGroup.HandleFunc("GET", "/getStatsForLeaderboardByProfileName", leaderboard.GetStatsForLeaderboardByProfileName)
	LeaderboardGroup.HandleFunc("GET", "/getPartyStat", leaderboard.GetPartyStat)
	if gameId == common.GameAoE3 {
		LeaderboardGroup.HandleFunc("GET", "/getAvatarStatLeaderBoard", leaderboard.GetAvatarStatLeaderBoard)
	}
	leaderboardGroup := gameGroup.Subgroup("/leaderboard")
	leaderboardGroup.HandleFunc("POST", "/applyOfflineUpdates", leaderboard.ApplyOfflineUpdates)
	leaderboardGroup.HandleFunc("POST", "/setAvatarStatValues", leaderboard.SetAvatarStatValues)

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
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		advertisementGroup.HandleFunc("POST", "/updatePlatformSessionID", advertisement.UpdatePlatformSessionID)
	}
	advertisementGroup.HandleFunc("POST", "/join", advertisement.Join)
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		advertisementGroup.HandleFunc("POST", "/updateTags", advertisement.UpdateTags)
	}
	advertisementGroup.HandleFunc("POST", "/update", advertisement.Update)
	advertisementGroup.HandleFunc("POST", "/leave", advertisement.Leave)
	advertisementGroup.HandleFunc("POST", "/host", advertisement.Host)
	if gameId == common.GameAoE1 || gameId == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/getLanAdvertisements", advertisement.GetLanAdvertisements)
	}
	if gameId == common.GameAoE2 {
		advertisementGroup.HandleFunc("GET", "/getLanAdvertisements", advertisement.GetLanAdvertisements)
	}
	if gameId == common.GameAoE1 || gameId == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/updatePlatformLobbyID", advertisement.UpdatePlatformLobbyID)
	}
	if gameId == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		advertisementGroup.HandleFunc("GET", "/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
	}
	advertisementGroup.HandleFunc("GET", "/getAdvertisements", advertisement.GetAdvertisements)
	if gameId == common.GameAoE1 || gameId == common.GameAoE3 {
		advertisementGroup.HandleFunc("POST", "/findAdvertisements", advertisement.FindAdvertisements)
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		advertisementGroup.HandleFunc("GET", "/findAdvertisements", advertisement.FindAdvertisements)
	}

	advertisementGroup.HandleFunc("POST", "/updateState", advertisement.UpdateState)

	if gameId == common.GameAoE2 || gameId == common.GameAoE3 || gameId == common.GameAoM {
		advertisementGroup.HandleFunc("POST", "/startObserving", advertisement.StartObserving)
		advertisementGroup.HandleFunc("POST", "/stopObserving", advertisement.StopObserving)
	}

	chatGroup := gameGroup.Subgroup("/chat")
	if gameId == common.GameAoE1 || gameId == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/getChatChannels", chat.GetChatChannels)
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		chatGroup.HandleFunc("GET", "/getChatChannels", chat.GetChatChannels)
	}
	chatGroup.HandleFunc("GET", "/getOfflineMessages", chat.GetOfflineMessages)
	if gameId == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/joinChannel", chat.JoinChannel)
	}
	if gameId == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/leaveChannel", chat.LeaveChannel)
	}
	if gameId == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/sendText", chat.SendText)
	}
	if gameId == common.GameAoE3 {
		chatGroup.HandleFunc("POST", "/sendWhisper", chat.SendWhisper)
	}
	if gameId == common.GameAoM {
		chatGroup.HandleFunc("POST", "/sendWhispers", chat.SendWhisper)
	}
	if gameId == common.GameAoM {
		chatGroup.HandleFunc("POST", "/deleteOfflineMessage", chat.DeleteOfflineMessage)
	}

	relationshipGroup := gameGroup.Subgroup("/relationship")
	if gameId == common.GameAoE1 || gameId == common.GameAoE3 {
		relationshipGroup.HandleFunc("POST", "/getRelationships", relationship.GetRelationships)
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		relationshipGroup.HandleFunc("GET", "/getRelationships", relationship.GetRelationships)
	}
	relationshipGroup.HandleFunc("GET", "/getPresenceData", relationship.GetPresenceData)
	relationshipGroup.HandleFunc("POST", "/setPresence", relationship.SetPresence)
	if gameId == common.GameAoE3 || gameId == common.GameAoM {
		relationshipGroup.HandleFunc("POST", "/setPresenceProperty", relationship.SetPresenceProperty)
	}
	if gameId == common.GameAoE3 || gameId == common.GameAoM {
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
	if gameId == common.GameAoM {
		partyGroup.HandleFunc("POST", "/createOrReportSinglePlayer", party.CreateOrReportSinglePlayer)
	}
	playerReportGroup := gameGroup.Subgroup("/playerreport")
	// TODO: Check if it applies to AoE I/AoE III
	if gameId == common.GameAoE2 || gameId == common.GameAoM {
		playerReportGroup.HandleFunc("POST", "/reportUser", playerreport.ReportUser)
	}

	invitationGroup := gameGroup.Subgroup("/invitation")
	invitationGroup.HandleFunc("POST", "/extendInvitation", invitation.ExtendInvitation)
	invitationGroup.HandleFunc("POST", "/cancelInvitation", invitation.CancelInvitation)
	invitationGroup.HandleFunc("POST", "/replyToInvitation", invitation.ReplyToInvitation)

	cloudGroup := gameGroup.Subgroup("/cloud")
	if gameId == common.GameAoE3 {
		cloudGroup.HandleFunc("POST", "/getFileURL", cloud.GetFileURL)
	}
	// TODO: Enable to AoM if/when it gets cloud support
	if gameId == common.GameAoE2 {
		cloudGroup.HandleFunc("GET", "/getFileURL", cloud.GetFileURL)
	}

	cloudGroup.HandleFunc("GET", "/getTempCredentials", cloud.GetTempCredentials)

	msstoreGroup := gameGroup.Subgroup("/msstore")
	msstoreGroup.HandleFunc("GET", "/getStoreTokens", msstore.GetStoreTokens)

	g.group.HandleFunc("GET", "/wss/", wss.Handle)

	// Artificial path needed for cloudGroup
	// TODO: Enable to AoM if/when it gets cloud support
	if gameId == common.GameAoE2 || gameId == common.GameAoE3 {
		g.group.HandleFunc("GET", "/cloudfiles/", cloudfiles.Cloudfiles)
	}
	return SessionMiddleware(g.group.mux)
}
