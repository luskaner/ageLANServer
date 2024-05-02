package wss

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/keyLock"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var timeoutTime = 60 * time.Minute

var lock = keyLock.NewKeyRWMutex()
var connections = make(map[string]*websocket.Conn)

func closeConn(conn *websocket.Conn, closeCode int, text string) {
	message := websocket.FormatCloseMessage(closeCode, text)
	if err := conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second)); err != nil {
		log.Println(err)
	}
	if err := conn.Close(); err != nil {
		log.Println(err)
	}
}

func parseMessage(message gin.H, currentSession *session.Info) (bool, uint32, *session.Info) {
	var sess *session.Info
	sess = nil
	op := uint32(message["operation"].(float64))
	if op == 0 {
		sessionToken, ok := message["sessionToken"]
		if ok {
			sess, ok = session.GetById(sessionToken.(string))
			if ok {
				return true, 0, sess
			} else {
				return false, 0, nil
			}
		}
	}
	if currentSession != nil {
		sess, _ = session.GetById(currentSession.GetId())
	}
	return false, op, sess
}

func Handle(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	done := make(chan struct{})
	timeout := time.AfterFunc(time.Minute, func() {
		select {
		case <-done:
		default:
			closeConn(conn, websocket.CloseNormalClosure, "Timeout")
		}
	})

	var msg gin.H
	err = conn.ReadJSON(&msg)

	if err != nil {
		defer func() {
			close(done)
			closeConn(conn, websocket.CloseNormalClosure, "Invalid login message format")
		}()
		return
	}
	reset, _, sess := parseMessage(msg, nil)
	if reset {
		timeout.Reset(timeoutTime)
	} else {
		defer func() {
			close(done)
			closeConn(conn, websocket.CloseNormalClosure, "Invalid login message data")
		}()
		return
	}

	sessionToken := sess.GetId()
	lock.Lock(sessionToken)
	connections[sessionToken] = conn
	lock.Unlock(sessionToken)

	for {
		err = conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		log.Printf("Received: %v", msg)
		reset, op, sess := parseMessage(msg, sess)
		if op == 0 {
			if sess == nil {
				break
			} else if sess.GetId() != sessionToken {
				lock.Lock(sessionToken)
				delete(connections, sessionToken)
				lock.Unlock(sessionToken)
				sessionToken = sess.GetId()
				lock.Lock(sessionToken)
				connections[sessionToken] = conn
				lock.Unlock(sessionToken)
			}
			if reset {
				timeout.Reset(timeoutTime)
			}
		} else if _, ok := session.GetById(sessionToken); !ok {
			break
		} else {
			// TODO: Handle other operations
		}
		log.Printf("Operation: %v", op)
	}

	lock.Lock(sessionToken)
	delete(connections, sessionToken)
	close(done)
	closeConn(conn, websocket.CloseNormalClosure, "Invalid message")
}

func SendMessage(sessionId string, message j.A) bool {
	lock.RLock(sessionId)

	conn, ok := connections[sessionId]

	if !ok {
		lock.RUnlock(sessionId)
		return false
	}

	err := conn.WriteJSON(message)
	lock.RUnlock(sessionId)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
