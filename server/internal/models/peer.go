package models

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"sync"
)

type MainPeer struct {
	advertisement *MainAdvertisement
	user          *MainUser
	race          int32
	team          int32
	invites       *i.SafeSet[*MainUser]
	lock          *sync.RWMutex
}

func (peer *MainPeer) GetAdvertisementId() int32 {
	return peer.advertisement.GetId()
}

func (peer *MainPeer) GetUser() *MainUser {
	return peer.user
}

func (peer *MainPeer) GetRace() int32 {
	peer.lock.RLock()
	defer peer.lock.RUnlock()
	return peer.race
}

func (peer *MainPeer) GetTeam() int32 {
	peer.lock.RLock()
	defer peer.lock.RUnlock()
	return peer.team
}

func (peer *MainPeer) Encode() i.A {
	peer.lock.RLock()
	defer peer.lock.RUnlock()
	return i.A{
		peer.advertisement.GetId(),
		peer.user.GetId(),
		-1,
		peer.user.GetStatId(),
		peer.race,
		peer.team,
		peer.advertisement.GetIp(),
	}
}

func (peer *MainPeer) Invite(user *MainUser) {
	peer.invites.Add(user)
}

func (peer *MainPeer) Uninvite(user *MainUser) {
	peer.invites.Delete(user)
}

func (peer *MainPeer) IsInvited(user *MainUser) bool {
	return peer.invites.Has(user)
}

func (peer *MainPeer) Update(race int32, team int32) {
	peer.lock.Lock()
	defer peer.lock.Unlock()
	peer.race = race
	peer.team = team
}
