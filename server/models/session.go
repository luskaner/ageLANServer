package models

import (
	"github.com/luskaner/aoe2DELanServer/server/internal"
	"sync"
	"time"
)

type Session struct {
	id     string
	expiry time.Time
	user   *User
}

var userIdSession sync.Map
var sessionStore sync.Map

var (
	sessionLetters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

func generateSessionId() string {
	for {
		sessionId := make([]rune, 30)
		for i := range sessionId {
			sessionId[i] = sessionLetters[internal.Rng.Intn(len(sessionLetters))]
		}
		sessionIdStr := string(sessionId)
		if _, exists := GetSessionById(sessionIdStr); !exists {
			return sessionIdStr
		}
	}
}

func (sess *Session) GetId() string {
	return sess.id
}

func (sess *Session) GetUser() *User {
	return sess.user
}

func CreateSession(user *User) string {
	removeSessionsExpired()
	session := &Session{
		id:     generateSessionId(),
		user:   user,
		expiry: time.Now().UTC().Add(time.Hour),
	}
	sessionStore.Store(session.id, session)
	userIdSession.Store(user.GetId(), session)
	return session.id
}

func (sess *Session) Delete() {
	userIdSession.Delete(sess.user.GetId())
	sessionStore.Delete(sess.id)
}

func GetSessionById(sessionId string) (*Session, bool) {
	value, exists := sessionStore.Load(sessionId)
	if exists {
		session := value.(*Session)
		if session.Expired() {
			session.Delete()
			exists = false
			session = nil
		}
		return session, exists
	}
	return nil, false
}

func GetSessionByUser(user *User) (*Session, bool) {
	value, exists := userIdSession.Load(user.GetId())
	if exists {
		session := value.(*Session)
		if session.Expired() {
			session.Delete()
			exists = false
			session = nil
		}
		return session, exists
	}
	return nil, false
}

func (sess *Session) Expired() bool {
	return time.Now().UTC().After(sess.expiry)
}

func removeSessionsExpired() {
	sessionStore.Range(func(_, value interface{}) bool {
		info := value.(*Session)
		if info.Expired() {
			info.Delete()
		}
		return true
	})
}
