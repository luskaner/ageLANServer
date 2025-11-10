package models

import (
	"net/http"
	"sync"
	"time"

	"github.com/luskaner/ageLANServer/server/internal"
)

type Session struct {
	id               string
	clientLibVersion uint16
	expiryTimer      *time.Timer
	expiryTimerLock  sync.Mutex
	userId           int32
	gameId           string
	messageChan      chan internal.A
}

var sessionStore = internal.NewSafeMap[string, *Session]()

var (
	sessionLetters  = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	sessionDuration = 5 * time.Minute
)

func generateSessionId() string {
	sessionId := make([]rune, 30)
	internal.WithRng(func(rand *internal.RandReader) {
		for i := range sessionId {
			sessionId[i] = sessionLetters[rand.IntN(len(sessionLetters))]
		}
	})
	return string(sessionId)
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

func (sess *Session) GetClientLibVersion() uint16 {
	return sess.clientLibVersion
}

func (sess *Session) AddMessage(message internal.A) {
	sess.messageChan <- message
}

func (sess *Session) WaitForMessages(ackNum uint) (uint, []internal.A) {
	var results []internal.A
	timer := time.NewTimer(19 * time.Second)
	defer timer.Stop()

	for {
		select {
		case msg := <-sess.messageChan:
			results = append(results, msg)
			for len(results) < cap(sess.messageChan) {
				select {
				case msg = <-sess.messageChan:
					results = append(results, msg)
				default:
					if len(results) > 0 {
						ackNum++
					}
					return ackNum, results
				}
			}
		case <-timer.C:
			return ackNum, results
		}
	}
}

func (sess *Session) Delete() {
	func() {
		sess.expiryTimerLock.Lock()
		defer sess.expiryTimerLock.Unlock()
		sess.expiryTimer.Stop()
	}()
	sessionStore.Delete(sess.id)
}

func (sess *Session) ResetExpiryTimer() {
	sess.expiryTimerLock.Lock()
	defer sess.expiryTimerLock.Unlock()
	sess.expiryTimer.Reset(sessionDuration)
}

func CreateSession(gameId string, userId int32, clientLibVersion uint16) string {
	sess := &Session{
		userId:           userId,
		gameId:           gameId,
		clientLibVersion: clientLibVersion,
		messageChan:      make(chan internal.A, 100),
	}
	defer func() {
		sess.expiryTimer = time.AfterFunc(sessionDuration, func() {
			sess.Delete()
		})
	}()
	for exists := true; exists; {
		sess.id = generateSessionId()
		_, exists = sessionStore.Store(sess.id, sess, func(_ *Session) bool {
			return false
		})
	}
	return sess.id
}

func GetSessionById(sessionId string) (*Session, bool) {
	return sessionStore.Load(sessionId)
}

func GetSessionByUserId(userId int32) (*Session, bool) {
	for sess := range sessionStore.Values() {
		if sess.userId == userId {
			return sess, true
		}
	}
	return nil, false
}

func SessionOrPanic(r *http.Request) *Session {
	sessAny, ok := session(r)
	if !ok {
		panic("Session should have been set already")
	}
	return sessAny
}

func session(r *http.Request) (*Session, bool) {
	sess, ok := r.Context().Value("session").(*Session)
	return sess, ok
}
