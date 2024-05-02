package models

import (
	internal "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"strings"
	"sync"
	"time"
)

type Info struct {
	id     string
	expiry time.Time
	user   *User
}

var userIdSession sync.Map
var sessionStore sync.Map
var anonymousPaths = map[string]bool{
	"/game/msstore/getStoreTokens": true,
	"/game/login/platformlogin":    true,
	"/wss/":                        true,
	"/game/news/getNews":           true,
}

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !anonymousPaths[c.Request.URL.Path] && !strings.HasPrefix(c.Request.URL.Path, "/cloudfiles/") {
			sessionID := c.Query("sessionID")
			if sessionID == "" {
				sessionID = c.PostForm("sessionID")
			}
			sess, ok := GetSessionById(sessionID)
			if !ok {
				c.JSON(401, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}
			c.Set("session", sess)
		}
		c.Next()
	}
}

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

func (sess *Info) GetId() string {
	return sess.id
}

func (sess *Info) GetUser() *User {
	return sess.user
}

func CreateSession(user *User) string {
	removeSessionsExpired()
	session := &Info{
		id:     generateSessionId(),
		user:   user,
		expiry: time.Now().UTC().Add(time.Hour),
	}
	sessionStore.Store(session.id, session)
	userIdSession.Store(user.GetId(), session)
	return session.id
}

func (sess *Info) Delete() {
	userIdSession.Delete(sess.user.GetId())
	sessionStore.Delete(sess.id)
}

func GetSessionById(sessionId string) (*Info, bool) {
	value, exists := sessionStore.Load(sessionId)
	if exists {
		session := value.(*Info)
		if session.Expired() {
			session.Delete()
			exists = false
			session = nil
		}
		return session, exists
	}
	return nil, false
}

func GetSessionByUser(user *User) (*Info, bool) {
	value, exists := userIdSession.Load(user.GetId())
	if exists {
		session := value.(*Info)
		if session.Expired() {
			session.Delete()
			exists = false
			session = nil
		}
		return session, exists
	}
	return nil, false
}

func (sess *Info) Expired() bool {
	return time.Now().UTC().After(sess.expiry)
}

func removeSessionsExpired() {
	sessionStore.Range(func(_, value interface{}) bool {
		info := value.(*Info)
		if info.Expired() {
			info.Delete()
		}
		return true
	})
}
