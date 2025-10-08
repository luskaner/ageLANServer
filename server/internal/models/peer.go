package models

import (
	"sync/atomic"

	i "github.com/luskaner/ageLANServer/server/internal"
)

type MainPeerMutable struct {
	Race int32
	Team int32
}

type MainPeer struct {
	advertisementId int32
	advertisementIp string
	userId          int32
	userStatId      int32
	mutable         *atomic.Value
	invites         *i.SafeSet[*MainUser]
}

func NewPeer(advertisementId int32, advertisementIp string, userId int32, userStatId int32, race int32, team int32) *MainPeer {
	peer := &MainPeer{
		advertisementId: advertisementId,
		advertisementIp: advertisementIp,
		userId:          userId,
		userStatId:      userStatId,
		mutable:         &atomic.Value{},
		invites:         i.NewSafeSet[*MainUser](),
	}
	peer.UpdateMutable(race, team)
	return peer
}

func (peer *MainPeer) GetMutable() *MainPeerMutable {
	return peer.mutable.Load().(*MainPeerMutable)
}

func (peer *MainPeer) Encode() i.A {
	mutable := peer.GetMutable()
	return i.A{
		peer.advertisementId,
		peer.userId,
		-1,
		peer.userStatId,
		mutable.Race,
		mutable.Team,
		peer.advertisementIp,
	}
}

func (peer *MainPeer) Invite(user *MainUser) bool {
	return peer.invites.Store(user)
}

func (peer *MainPeer) Uninvite(user *MainUser) bool {
	return peer.invites.Delete(user)
}

func (peer *MainPeer) UpdateMutable(race int32, team int32) {
	peer.mutable.Store(&MainPeerMutable{race, team})
}
