package routes

import (
	"aoe2DELanServer/routes/cloudfiles"
	achievement "aoe2DELanServer/routes/game/Achievement"
	"aoe2DELanServer/routes/game/Challenge"
	"aoe2DELanServer/routes/game/CommunityEvent"
	"aoe2DELanServer/routes/game/Leaderboard"
	"aoe2DELanServer/routes/game/account"
	"aoe2DELanServer/routes/game/advertisement"
	automatch2 "aoe2DELanServer/routes/game/automatch2"
	"aoe2DELanServer/routes/game/chat"
	"aoe2DELanServer/routes/game/clan"
	"aoe2DELanServer/routes/game/cloud"
	"aoe2DELanServer/routes/game/invitation"
	"aoe2DELanServer/routes/game/item"
	"aoe2DELanServer/routes/game/login"
	"aoe2DELanServer/routes/game/news"
	"aoe2DELanServer/routes/game/party"
	"aoe2DELanServer/routes/game/relationship"
	"aoe2DELanServer/routes/msstore"
	"aoe2DELanServer/routes/wss"
	"github.com/gin-gonic/gin"
)

func Initialize(r *gin.Engine) {
	gameGroup := r.Group("/game")
	{
		itemGroup := gameGroup.Group("/item")
		{
			itemGroup.GET("/getItemBundleItemsJson", item.GetItemBundleItemsJson)
			itemGroup.GET("/getItemDefinitionsJson", item.GetItemDefinitionsJson)
			itemGroup.GET("/getItemLoadouts", item.GetItemLoadouts)
			itemGroup.POST("/signItems", item.SignItems)
			itemGroup.GET("/getInventoryByProfileIDs", item.GetInventoryByProfileIDs)
		}
		clanGroup := gameGroup.Group("/clan")
		{
			clanGroup.POST("/create", clan.Create)
			clanGroup.GET("/find", clan.Find)
		}
		communityEventGroup := gameGroup.Group("/CommunityEvent")
		{
			communityEventGroup.GET("/getAvailableCommunityEvents", communityEvent.GetAvailableCommunityEvents)
		}
		challengeGroup := gameGroup.Group("/Challenge")
		{
			challengeGroup.GET("/getChallengeProgress", challenge.GetChallengeProgress)
			challengeGroup.GET("/getChallenges", challenge.GetChallenges)
		}
		newsGroup := gameGroup.Group("/news")
		{
			newsGroup.GET("/getNews", news.GetNews)
		}
		loginGroup := gameGroup.Group("/login")
		{
			loginGroup.POST("/platformlogin", login.Platformlogin)
			loginGroup.POST("/logout", login.Logout)
		}
		accountGroup := gameGroup.Group("/account")
		{
			accountGroup.POST("/setLanguage", account.SetLanguage)
			accountGroup.POST("/setCrossplayEnabled", account.SetCrossplayEnabled)
			accountGroup.POST("/setAvatarMetadata", account.SetAvatarMetadata)
			accountGroup.POST("/FindProfilesByPlatformID", account.FindProfilesByPlatformID)
			accountGroup.GET("/FindProfiles", account.FindProfiles)
			accountGroup.GET("/getProfileName", account.GetProfileName)
		}
		LeaderboardGroup := gameGroup.Group("/Leaderboard")
		{
			LeaderboardGroup.POST("/applyOfflineUpdates", leaderboard.ApplyOfflineUpdates)
			LeaderboardGroup.GET("/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
			LeaderboardGroup.GET("/getLeaderBoard", leaderboard.GetLeaderBoard)
			LeaderboardGroup.GET("/getAvailableLeaderboards", leaderboard.GetAvailableLeaderboards)
			LeaderboardGroup.GET("/getStatGroupsByProfileIDs", leaderboard.GetStatGroupsByProfileIDs)
			LeaderboardGroup.GET("/getStatsForLeaderboardByProfileName", leaderboard.GetStatsForLeaderboardByProfileName)
			LeaderboardGroup.GET("/getPartyStat", leaderboard.GetPartyStat)
		}
		leaderboardGroup := gameGroup.Group("/leaderboard")
		{
			leaderboardGroup.POST("/setAvatarStatValues", leaderboard.SetAvatarStatValues)
		}
		automatch2Group := gameGroup.Group("/automatch2")
		{
			automatch2Group.GET("/getAutomatchMap", automatch2.GetAutomatchMap)
		}
		AchievementGroup := gameGroup.Group("/Achievement")
		{
			AchievementGroup.GET("/getAchievements", achievement.GetAchievements)
			AchievementGroup.GET("/getAvailableAchievements", achievement.GetAvailableAchievements)
		}
		achievementGroup := gameGroup.Group("/achievement")
		{
			achievementGroup.POST("/grantAchievement", achievement.GrantAchievement)
			achievementGroup.POST("/syncStats", achievement.SyncStats)
		}
		advertisementGroup := gameGroup.Group("/advertisement")
		{
			advertisementGroup.POST("/updatePlatformSessionID", advertisement.UpdatePlatformSessionID)
			advertisementGroup.POST("/join", advertisement.Join)
			advertisementGroup.POST("/updateTags", advertisement.UpdateTags)
			advertisementGroup.POST("/update", advertisement.Update)
			advertisementGroup.POST("/leave", advertisement.Leave)
			advertisementGroup.POST("/host", advertisement.Host)
			advertisementGroup.GET("/getLanAdvertisements", advertisement.GetLanAdvertisements)
			advertisementGroup.GET("/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
			advertisementGroup.GET("/getAdvertisements", advertisement.GetAdvertisements)
			advertisementGroup.GET("/findAdvertisements", advertisement.FindAdvertisements)
			advertisementGroup.POST("/updateState", advertisement.UpdateState)
		}
		chatGroup := gameGroup.Group("/chat")
		{
			chatGroup.GET("/getChatChannels", chat.GetChatChannels)
			chatGroup.GET("/getOfflineMessages", chat.GetOfflineMessages)
		}
		relationshipGroup := gameGroup.Group("/relationship")
		{
			relationshipGroup.GET("/getRelationships", relationship.GetRelationships)
			relationshipGroup.GET("/getPresenceData", relationship.GetPresenceData)
			relationshipGroup.POST("/setPresence", relationship.SetPresence)
			relationshipGroup.POST("/ignore", relationship.Ignore)
			relationshipGroup.POST("/clearRelationship", relationship.ClearRelationship)
		}
		partyGroup := gameGroup.Group("/party")
		{
			partyGroup.POST("/peerAdd", party.PeerAdd)
			partyGroup.POST("/peerUpdate", party.PeerUpdate)
			partyGroup.POST("/sendMatchChat", party.SendMatchChat)
			partyGroup.POST("/reportMatch", party.ReportMatch)
			partyGroup.POST("/finalizeReplayUpload", party.FinalizeReplayUpload)
			partyGroup.POST("/updateHost", party.UpdateHost)
		}
		invitationGroup := gameGroup.Group("/invitation")
		{
			invitationGroup.POST("/extendInvitation", invitation.ExtendInvitation)
			invitationGroup.POST("/cancelInvitation", invitation.CancelInvitation)
			invitationGroup.POST("/replyToInvitation", invitation.ReplyToInvitation)
		}
		cloudGroup := gameGroup.Group("/cloud")
		{
			cloudGroup.GET("/getFileURL", cloud.GetFileURL)
			cloudGroup.GET("/getTempCredentials", cloud.GetTempCredentials)
		}
		msstoreGroup := gameGroup.Group("/msstore")
		{
			msstoreGroup.GET("/getStoreTokens", msstore.GetStoreTokens)
		}
	}
	r.GET("/wss/", wss.Handle)
	r.GET("/cloudfiles/*key", cloudfiles.Cloudfiles)
}
