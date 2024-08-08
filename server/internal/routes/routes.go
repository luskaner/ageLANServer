package routes

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/cloudfiles"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/account"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/achievement"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement"
	Automatch2 "github.com/luskaner/aoe2DELanServer/server/internal/routes/game/automatch2"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/challenge"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/chat"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/clan"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/cloud"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/communityEvent"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/invitation"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/item"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/leaderboard"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/login"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/news"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/party"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/relationship"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/msstore"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/test"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
)

type Group struct {
	parent *Group
	path   string
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
	}
}

func (g *Group) HandleFunc(mux *http.ServeMux, method string, path string, handler http.HandlerFunc) {
	mux.HandleFunc(method+" "+g.fullPath()+path, handler)
}

func Initialize(mux *http.ServeMux) {
	baseGroup := Group{
		path: "",
	}
	gameGroup := baseGroup.Subgroup("/game")
	itemGroup := gameGroup.Subgroup("/item")
	itemGroup.HandleFunc(mux, "GET", "/getItemBundleItemsJson", item.GetItemBundleItemsJson)
	itemGroup.HandleFunc(mux, "GET", "/getItemDefinitionsJson", item.GetItemDefinitionsJson)
	itemGroup.HandleFunc(mux, "GET", "/getItemLoadouts", item.GetItemLoadouts)
	itemGroup.HandleFunc(mux, "POST", "/signItems", item.SignItems)
	itemGroup.HandleFunc(mux, "GET", "/getInventoryByProfileIDs", item.GetInventoryByProfileIDs)

	clanGroup := gameGroup.Subgroup("/clan")
	clanGroup.HandleFunc(mux, "POST", "/create", clan.Create)
	clanGroup.HandleFunc(mux, "GET", "/find", clan.Find)

	communityEventGroup := gameGroup.Subgroup("/CommunityEvent")
	communityEventGroup.HandleFunc(mux, "GET", "/getAvailableCommunityEvents", communityEvent.GetAvailableCommunityEvents)

	challengeGroup := gameGroup.Subgroup("/Challenge")
	challengeGroup.HandleFunc(mux, "GET", "/getChallengeProgress", challenge.GetChallengeProgress)
	challengeGroup.HandleFunc(mux, "GET", "/getChallenges", challenge.GetChallenges)

	newsGroup := gameGroup.Subgroup("/news")
	newsGroup.HandleFunc(mux, "GET", "/getNews", news.GetNews)

	loginGroup := gameGroup.Subgroup("/login")
	loginGroup.HandleFunc(mux, "POST", "/platformlogin", login.Platformlogin)
	loginGroup.HandleFunc(mux, "POST", "/logout", login.Logout)

	accountGroup := gameGroup.Subgroup("/account")
	accountGroup.HandleFunc(mux, "POST", "/setLanguage", account.SetLanguage)
	accountGroup.HandleFunc(mux, "POST", "/setCrossplayEnabled", account.SetCrossplayEnabled)
	accountGroup.HandleFunc(mux, "POST", "/setAvatarMetadata", account.SetAvatarMetadata)
	accountGroup.HandleFunc(mux, "POST", "/FindProfilesByPlatformID", account.FindProfilesByPlatformID)
	accountGroup.HandleFunc(mux, "GET", "/FindProfiles", account.FindProfiles)
	accountGroup.HandleFunc(mux, "GET", "/getProfileName", account.GetProfileName)

	LeaderboardGroup := gameGroup.Subgroup("/Leaderboard")
	LeaderboardGroup.HandleFunc(mux, "GET", "/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
	LeaderboardGroup.HandleFunc(mux, "GET", "/getLeaderBoard", leaderboard.GetLeaderBoard)
	LeaderboardGroup.HandleFunc(mux, "GET", "/getAvailableLeaderboards", leaderboard.GetAvailableLeaderboards)
	LeaderboardGroup.HandleFunc(mux, "GET", "/getStatGroupsByProfileIDs", leaderboard.GetStatGroupsByProfileIDs)
	LeaderboardGroup.HandleFunc(mux, "GET", "/getStatsForLeaderboardByProfileName", leaderboard.GetStatsForLeaderboardByProfileName)
	LeaderboardGroup.HandleFunc(mux, "GET", "/getPartyStat", leaderboard.GetPartyStat)

	leaderboardGroup := gameGroup.Subgroup("/leaderboard")
	leaderboardGroup.HandleFunc(mux, "POST", "/applyOfflineUpdates", leaderboard.ApplyOfflineUpdates)
	leaderboardGroup.HandleFunc(mux, "POST", "/setAvatarStatValues", leaderboard.SetAvatarStatValues)

	automatch2Group := gameGroup.Subgroup("/automatch2")
	automatch2Group.HandleFunc(mux, "GET", "/getAutomatchMap", Automatch2.GetAutomatchMap)

	AchievementGroup := gameGroup.Subgroup("/Achievement")
	AchievementGroup.HandleFunc(mux, "GET", "/getAchievements", achievement.GetAchievements)
	AchievementGroup.HandleFunc(mux, "GET", "/getAvailableAchievements", achievement.GetAvailableAchievements)

	achievementGroup := gameGroup.Subgroup("/achievement")
	achievementGroup.HandleFunc(mux, "POST", "/applyOfflineUpdates", achievement.ApplyOfflineUpdates)
	achievementGroup.HandleFunc(mux, "POST", "/grantAchievement", achievement.GrantAchievement)
	achievementGroup.HandleFunc(mux, "POST", "/syncStats", achievement.SyncStats)

	advertisementGroup := gameGroup.Subgroup("/advertisement")
	advertisementGroup.HandleFunc(mux, "POST", "/updatePlatformSessionID", advertisement.UpdatePlatformSessionID)
	advertisementGroup.HandleFunc(mux, "POST", "/join", advertisement.Join)
	advertisementGroup.HandleFunc(mux, "POST", "/updateTags", advertisement.UpdateTags)
	advertisementGroup.HandleFunc(mux, "POST", "/update", advertisement.Update)
	advertisementGroup.HandleFunc(mux, "POST", "/leave", advertisement.Leave)
	advertisementGroup.HandleFunc(mux, "POST", "/host", advertisement.Host)
	advertisementGroup.HandleFunc(mux, "GET", "/getLanAdvertisements", advertisement.GetLanAdvertisements)
	advertisementGroup.HandleFunc(mux, "GET", "/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
	advertisementGroup.HandleFunc(mux, "GET", "/getAdvertisements", advertisement.GetAdvertisements)
	advertisementGroup.HandleFunc(mux, "GET", "/findAdvertisements", advertisement.FindAdvertisements)
	advertisementGroup.HandleFunc(mux, "POST", "/updateState", advertisement.UpdateState)

	chatGroup := gameGroup.Subgroup("/chat")
	chatGroup.HandleFunc(mux, "GET", "/getChatChannels", chat.GetChatChannels)
	chatGroup.HandleFunc(mux, "GET", "/getOfflineMessages", chat.GetOfflineMessages)

	relationshipGroup := gameGroup.Subgroup("/relationship")
	relationshipGroup.HandleFunc(mux, "GET", "/getRelationships", relationship.GetRelationships)
	relationshipGroup.HandleFunc(mux, "GET", "/getPresenceData", relationship.GetPresenceData)
	relationshipGroup.HandleFunc(mux, "POST", "/setPresence", relationship.SetPresence)
	relationshipGroup.HandleFunc(mux, "POST", "/ignore", relationship.Ignore)
	relationshipGroup.HandleFunc(mux, "POST", "/clearRelationship", relationship.ClearRelationship)

	partyGroup := gameGroup.Subgroup("/party")
	partyGroup.HandleFunc(mux, "POST", "/peerAdd", party.PeerAdd)
	partyGroup.HandleFunc(mux, "POST", "/peerUpdate", party.PeerUpdate)
	partyGroup.HandleFunc(mux, "POST", "/sendMatchChat", party.SendMatchChat)
	partyGroup.HandleFunc(mux, "POST", "/reportMatch", party.ReportMatch)
	partyGroup.HandleFunc(mux, "POST", "/finalizeReplayUpload", party.FinalizeReplayUpload)
	partyGroup.HandleFunc(mux, "POST", "/updateHost", party.UpdateHost)

	invitationGroup := gameGroup.Subgroup("/invitation")
	invitationGroup.HandleFunc(mux, "POST", "/extendInvitation", invitation.ExtendInvitation)
	invitationGroup.HandleFunc(mux, "POST", "/cancelInvitation", invitation.CancelInvitation)
	invitationGroup.HandleFunc(mux, "POST", "/replyToInvitation", invitation.ReplyToInvitation)

	cloudGroup := gameGroup.Subgroup("/cloud")
	cloudGroup.HandleFunc(mux, "GET", "/getFileURL", cloud.GetFileURL)
	cloudGroup.HandleFunc(mux, "GET", "/getTempCredentials", cloud.GetTempCredentials)

	msstoreGroup := gameGroup.Subgroup("/msstore")
	msstoreGroup.HandleFunc(mux, "GET", "/getStoreTokens", msstore.GetStoreTokens)

	// Used for the launcher
	baseGroup.HandleFunc(mux, "GET", "/test", test.Test)
	baseGroup.HandleFunc(mux, "GET", "/wss/", wss.Handle)
	baseGroup.HandleFunc(mux, "GET", "/cloudfiles/", cloudfiles.Cloudfiles)
}
