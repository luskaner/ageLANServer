package models

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
)

type Message struct {
	advertisementId int32
	time            int64
	broadcast       bool
	content         string
	typ             uint8
	sender          *User
	receivers       []*User
}

func (message *Message) GetAdvertisement() *Advertisement {
	adv, _ := GetAdvertisement(message.advertisementId)
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

func (message *Message) GetSender() *User {
	return message.sender
}

func (message *Message) GetReceivers() []*User {
	return message.receivers
}

func (message *Message) Encode() i.A {
	return i.A{
		message.sender.GetId(),
		message.content,
		message.content,
		message.typ,
		message.advertisementId,
	}
}
