package fiber_session_middleware

import (
	"github.com/gofiber/fiber"
	fbSession "github.com/gofiber/session"
)

const SessionLocalKey = "session"

type sessionMiddleware struct {
	sessionContainer *fbSession.Session
}

type Session interface {
	ID() string
	Set(key string, value interface{})
	Get(key string) interface{}
	Delete(key string)
	Destroy()
}

type session struct {
	*fbSession.Store
	changed bool
}

func (s *session) Set(key string, value interface{}) {
	s.changed = true
	s.Store.Set(key, value)
}

func (s *session) Delete(key string) {
	s.changed = true
	s.Store.Delete(key)
}

func (mw *sessionMiddleware) handler(c *fiber.Ctx) {
	store := mw.sessionContainer.Get(c)
	defer func() {
		if l := c.Locals(SessionLocalKey); l != nil {
			s := l.(*session)
			if s.changed {
				store.Save()
			}
		}
	}()
	c.Locals(SessionLocalKey, &session{Store: store})
	c.Next()
}

func New(config ...fbSession.Config) func(*fiber.Ctx) {
	var sessionContainer *fbSession.Session
	if len(config) == 0 {
		sessionContainer = fbSession.New()
	} else {
		sessionContainer = fbSession.New(config[0])
	}
	mw := &sessionMiddleware{sessionContainer: sessionContainer}
	return mw.handler
}
