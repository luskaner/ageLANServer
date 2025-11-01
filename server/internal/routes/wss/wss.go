package wss

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type connectionWrapper struct {
	writeLock *sync.Mutex
	conn      *websocket.Conn
}

func (c *connectionWrapper) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

func (c *connectionWrapper) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *connectionWrapper) WriteJSON(v any) error {
	if err := c.conn.WriteJSON(v); err == nil {
		c.logJSON(c.LocalAddr(), c.RemoteAddr(), v)
		return nil
	} else {
		return err
	}
}

func (c *connectionWrapper) ReadJSON(v any) error {
	if err := c.conn.ReadJSON(v); err == nil {
		c.logJSON(c.RemoteAddr(), c.LocalAddr(), v)
		return nil
	} else {
		return err
	}
}

func (c *connectionWrapper) logJSON(sender string, receiver string, data any) {
	logger.LogMessage(
		"wss",
		slog.String("type", "json"),
		slog.String("sender", sender),
		slog.String("receiver", receiver),
		slog.Any("data", data),
	)
}

func (c *connectionWrapper) logControl(sender string, receiver string, messageType int, data []byte) {
	bodyHash := ""
	if len(data) > 0 {
		hash := sha512.Sum512(data)
		bodyHash = base64.StdEncoding.EncodeToString(hash[:])
	}
	logger.LogMessage(
		"wss",
		slog.String("type", "control"),
		slog.String("sender", sender),
		slog.String("receiver", receiver),
		slog.Group("data",
			slog.Int("messageType", messageType),
			slog.String("body", base64.StdEncoding.EncodeToString(data)),
			slog.String("body_hash", bodyHash),
		),
	)
}

func (c *connectionWrapper) WriteControl(messageType int, data []byte, deadline time.Time) error {
	c.logControl(c.LocalAddr(), c.RemoteAddr(), messageType, data)
	return c.conn.WriteControl(messageType, data, deadline)
}

func (c *connectionWrapper) logClose(sender string, receiver string) {
	logger.LogMessage(
		"wss",
		slog.String("type", "disconnection"),
		slog.String("sender", sender),
		slog.String("receiver", receiver),
	)
}

func (c *connectionWrapper) Close() error {
	defer func() {
		c.conn = nil
		c.writeLock = nil
	}()
	c.logClose(c.LocalAddr(), c.RemoteAddr())
	return c.conn.Close()
}

func (c *connectionWrapper) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *connectionWrapper) SetPingHandler(h func(appData string) error) {
	c.conn.SetPingHandler(func(appData string) error {
		c.logControl(c.RemoteAddr(), c.LocalAddr(), websocket.PingMessage, []byte(appData))
		return h(appData)
	})
}

func (c *connectionWrapper) SetCloseHandler(h func(code int, text string) error) {
	c.conn.SetCloseHandler(func(code int, text string) error {
		c.logClose(c.LocalAddr(), c.RemoteAddr())
		return h(code, text)
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var connections = i.NewSafeMap[string, *connectionWrapper]()
var writeWait = 1 * time.Second

func closeConn(conn *connectionWrapper, closeCode int, text string) {
	message := websocket.FormatCloseMessage(closeCode, text)
	_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
	_ = conn.Close()
}

func parseMessage(message i.H, currentSession *models.Session) (uint32, *models.Session) {
	var sess *models.Session
	sess = nil
	op := uint32(message["operation"].(float64))
	if op == 0 {
		sessionToken, ok := message["sessionToken"]
		if ok {
			sess, ok = models.GetSessionById(sessionToken.(string))
			if ok {
				return 0, sess
			} else {
				return 0, nil
			}
		}
	}
	if currentSession != nil {
		var ok bool
		sess, ok = models.GetSessionById(currentSession.GetId())
		if !ok {
			return 0, nil
		}
	}
	return op, sess
}

func Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	connWrapper := &connectionWrapper{
		writeLock: &sync.Mutex{},
		conn:      conn,
	}
	logger.LogMessage(
		"wss",
		slog.String("type", "connection"),
		slog.String("receiver", connWrapper.LocalAddr()),
		slog.String("sender", connWrapper.RemoteAddr()),
		slog.Group("data",
			slog.String("host", r.Host),
		),
	)
	connWrapper.conn.SetCloseHandler(func(code int, text string) error {
		closeConn(connWrapper, code, text)
		return nil
	})
	err = connWrapper.SetReadDeadline(time.Now().Add(time.Minute))
	if err != nil {
		closeConn(connWrapper, websocket.CloseNormalClosure, "Missing initial message")
		return
	}

	var loginMsg i.H
	err = connWrapper.ReadJSON(&loginMsg)

	if err != nil {
		closeConn(connWrapper, websocket.CloseNormalClosure, "Invalid or absent login message")
		return
	}

	_, sess := parseMessage(loginMsg, nil)
	if sess == nil {
		closeConn(connWrapper, websocket.CloseNormalClosure, "Invalid login message data")
		return
	}

	sessionToken := sess.GetId()

	connections.Store(
		sessionToken,
		connWrapper,
		nil,
	)

	sess.ResetExpiryTimer()

	connWrapper.SetPingHandler(func(message string) error {
		var pingErr error
		pingErr = connWrapper.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(writeWait))
		if pingErr == nil {
			pingErr = connWrapper.SetReadDeadline(time.Now().Add(time.Minute))
			if pingErr == nil {
				sess.ResetExpiryTimer()
			}
		} else if errors.Is(pingErr, websocket.ErrCloseSent) {
			return nil
		} else {
			var e net.Error
			if errors.As(pingErr, &e) && e.Temporary() {
				return nil
			}
		}
		return pingErr
	})

	defer func() {
		connections.Delete(sessionToken)
		closeConn(connWrapper, websocket.CloseNormalClosure, "Invalid message")
	}()

	var op uint32
	for {
		var msg i.H
		err = connWrapper.ReadJSON(&msg)
		if err != nil {
			break
		}
		op, sess = parseMessage(msg, sess)
		if op == 0 {
			if sess == nil {
				break
			}
			if sessId := sess.GetId(); sessId != sessionToken {
				connections.StoreAndDelete(sessId, connWrapper, sessionToken)
				sessionToken = sessId
			}
		} else if _, ok := models.GetSessionById(sessionToken); !ok {
			break
		}
		sess.ResetExpiryTimer()
	}
}

func sendMessage(sessionId string, message i.A) bool {
	connWrapper, ok := connections.Load(sessionId)

	if !ok {
		return false
	}

	connWrapper.writeLock.Lock()
	defer connWrapper.writeLock.Unlock()
	err := connWrapper.WriteJSON(message)

	if err != nil {
		return false
	}

	return true
}

func SendOrStoreMessage(session *models.Session, action string, message i.A) {
	finalMessage := i.A{0, action, session.GetUserId(), message}
	go func(session *models.Session, finalMessage i.A) {
		if ok := sendMessage(session.GetId(), finalMessage); !ok {
			session.AddMessage(finalMessage)
		}
	}(session, finalMessage)
}
