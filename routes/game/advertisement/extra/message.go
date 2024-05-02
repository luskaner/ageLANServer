package extra

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/user"
)

type Message struct {
	advertisementId uint32
	time            int64
	broadcast       bool
	content         string
	typ             uint8
	sender          *user.User
	receivers       []*user.User
}

func (message *Message) GetAdvertisement() *Advertisement {
	adv, _ := Get(message.advertisementId)
	return adv
}

func (message *Message) GetTime() int64 {
	return message.time
}

func (message *Message) GetBroadcast() bool {
	return message.broadcast
}

func (message *Message) GetContent() string {
	return message.content
}

func (message *Message) GetType() uint8 {
	return message.typ
}

func (message *Message) GetSender() *user.User {
	return message.sender
}

func (message *Message) GetReceivers() []*user.User {
	return message.receivers
}

func (message *Message) Encode() j.A {
	return j.A{
		message.sender.GetId(),
		message.content,
		message.content,
		message.typ,
		message.advertisementId,
	}
}
