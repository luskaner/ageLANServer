package models

import (
	"github.com/luskaner/ageLANServer/server/internal"
	"sync"
	"time"
)

type Session struct {
	id              string
	expiryTimer     *time.Timer
	userId          int32
	expiryTimerLock sync.Mutex
	gameId          string
	messages        []internal.A
	messagesLock    sync.RWMutex
	messagesIndex   uint8
}

var userIdSession = internal.NewSafeMap[int32, *Session]()
var sessionStore = internal.NewSafeMap[string, *Session]()

var (
	sessionLetters  = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	sessionDuration = 5 * time.Minute
)

func generateSessionId() string {
	sessionId := make([]rune, 30)
	for {
		for i := range sessionId {
			internal.RngLock.Lock()
			sessionId[i] = sessionLetters[internal.Rng.Intn(len(sessionLetters))]
			internal.RngLock.Unlock()
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

func (sess *Session) GetUserId() int32 {
	return sess.userId
}

func (sess *Session) GetGameId() string {
	return sess.gameId
}

func (sess *Session) AddMessage(message internal.A) {
	for {
		sess.messagesLock.RLock()
		if sess.messagesIndex == uint8(len(sess.messages)) {
			sess.messagesLock.RUnlock()
			time.Sleep(time.Millisecond * 100)
			continue
		}
		sess.messagesLock.RUnlock()
		sess.messagesLock.Lock()
		sess.messages[sess.messagesIndex] = message
		sess.messagesIndex++
		sess.messagesLock.Unlock()
		break
	}
}

func (sess *Session) WaitForMessages(ackNum uint) (uint, []internal.A) {
	results := make([]internal.A, 0)
	timer := time.NewTimer(time.Second * 19)
	defer timer.Stop()

	returnFn := func() (uint, []internal.A) {
		if len(results) > 0 {
			ackNum++
		}
		return ackNum, results
	}

	for {
		select {
		case <-timer.C:
			return returnFn()
		default:
			var i uint8
			sess.messagesLock.RLock()
			for i = 0; i < sess.messagesIndex; i++ {
				results = append(results, sess.messages[i])
			}
			sess.messagesLock.RUnlock()
			sess.messagesLock.Lock()
			sess.messagesIndex = 0
			sess.messagesLock.Unlock()
			if len(results) > 0 {
				return returnFn()
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func CreateSession(gameId string, userId int32) string {
	session := &Session{
		id:       generateSessionId(),
		userId:   userId,
		gameId:   gameId,
		messages: make([]internal.A, 10),
	}
	session.expiryTimer = time.AfterFunc(sessionDuration, func() {
		session.Delete()
	})
	sessionStore.Store(session.id, session)
	userIdSession.Store(userId, session)
	return session.id
}

func (sess *Session) Delete() {
	_ = sess.expiryTimer.Stop()
	userIdSession.Delete(sess.userId)
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
	return sessionStore.Load(sessionId)
}

func GetSessionByUserId(userId int32) (*Session, bool) {
	return userIdSession.Load(userId)
}
