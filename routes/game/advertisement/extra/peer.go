package extra

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/keyLock"
	"aoe2DELanServer/user"
)

type Peer struct {
	advertisement *Advertisement
	user          *user.User
	race          int32
	team          int32
	invites       map[*user.User]interface{}
}

var invitesLock = keyLock.NewKeyRWMutex()

func (peer *Peer) GetAdvertisement() *Advertisement {
	return peer.advertisement
}

func (peer *Peer) GetUser() *user.User {
	return peer.user
}

func (peer *Peer) GetRace() int32 {
	return peer.race
}

func (peer *Peer) GetTeam() int32 {
	return peer.team
}

func (peer *Peer) Encode() j.A {
	return j.A{
		peer.advertisement.GetId(),
		peer.user.GetId(),
		-1,
		peer.user.GetStatId(),
		peer.race,
		peer.team,
		peer.advertisement.ip,
	}
}

func (peer *Peer) Invite(user *user.User) {
	userId := user.GetId()
	invitesLock.Lock(userId)
	defer invitesLock.Unlock(userId)
	peer.invites[user] = struct{}{}
}

func (peer *Peer) Uninvite(user *user.User) {
	userId := user.GetId()
	invitesLock.Lock(userId)
	defer invitesLock.Unlock(userId)
	delete(peer.invites, user)
}

func (peer *Peer) IsInvited(user *user.User) bool {
	userId := user.GetId()
	invitesLock.RLock(userId)
	defer invitesLock.RUnlock(userId)
	_, ok := peer.invites[user]
	return ok
}
