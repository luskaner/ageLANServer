package models

import (
	"sync/atomic"

	i "github.com/luskaner/ageLANServer/server/internal"
)

type MainPeerMutable struct {
	Race int32
	Team int32
}

type Peer interface {
	GetUserId() int32
	GetParty() int32
	Encode() i.A
	Invite(user User) bool
	Uninvite(user User) bool
	GetMutable() *MainPeerMutable
	UpdateMutable(race int32, team int32)
}

type MainPeer struct {
	advertisementId int32
	party           int32
	advertisementIp string
	userId          int32
	userStatId      int32
	mutable         *atomic.Value
	invites         *i.SafeSet[User]
}

func NewPeer(advertisementId int32, advertisementIp string, userId int32, userStatId int32, party int32, race int32, team int32) Peer {
	peer := &MainPeer{
		advertisementId: advertisementId,
		party:           party,
		advertisementIp: advertisementIp,
		userId:          userId,
		userStatId:      userStatId,
		mutable:         &atomic.Value{},
		invites:         i.NewSafeSet[User](),
	}
	peer.UpdateMutable(race, team)
	return peer
}

func (peer *MainPeer) GetUserId() int32 {
	return peer.userId
}

func (peer *MainPeer) GetParty() int32 {
	return peer.party
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

func (peer *MainPeer) Invite(user User) bool {
	return peer.invites.Store(user)
}

func (peer *MainPeer) Uninvite(user User) bool {
	return peer.invites.Delete(user)
}

func (peer *MainPeer) UpdateMutable(race int32, team int32) {
	peer.mutable.Store(&MainPeerMutable{race, team})
}
