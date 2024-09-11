package models

import (
	"github.com/luskaner/aoe2DELanServer/server/internal"
	"sync"
	"time"
)

type Session struct {
	id              string
	expiryTimer     *time.Timer
	user            *User
	expiryTimerLock sync.Mutex
}

var userIdSession sync.Map
var sessionStore sync.Map

var (
	sessionLetters  = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	sessionDuration = 5 * time.Minute
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
	session := &Session{
		id:   generateSessionId(),
		user: user,
	}
	session.expiryTimer = time.AfterFunc(sessionDuration, func() {
		session.Delete()
	})
	sessionStore.Store(session.id, session)
	userIdSession.Store(user.GetId(), session)
	return session.id
}

func (sess *Session) Delete() {
	_ = sess.expiryTimer.Stop()
	userIdSession.Delete(sess.user.GetId())
	sessionStore.Delete(sess.id)
}

func (sess *Session) ResetExpiryTimer() {
	sess.expiryTimerLock.Lock()
	defer sess.expiryTimerLock.Unlock()
	if !sess.expiryTimer.Stop() {
		<-sess.expiryTimer.C
	}
	sess.expiryTimer.Reset(sessionDuration)
}

func GetSessionById(sessionId string) (*Session, bool) {
	value, exists := sessionStore.Load(sessionId)
	if exists {
		return value.(*Session), true
	}
	return nil, false
}

func GetSessionByUser(user *User) (*Session, bool) {
	value, exists := userIdSession.Load(user.GetId())
	if exists {
		return value.(*Session), true
	}
	return nil, false
}
