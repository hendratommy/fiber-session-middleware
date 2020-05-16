package fiber_session_middleware

import (
	"github.com/gofiber/fiber"
	fbSession "github.com/gofiber/session"
)

const SessionLocalKey = "session"

type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool
	// gofiber/session Config
	StoreConfig fbSession.Config
}

type sessionMiddleware struct {
	cfg Config
	core *fbSession.Session
}

// `ctx.Locals("session")` will return `Session` interface
type Session interface {
	// Return current session's id
	ID() string
	// Set the data to store
	Set(key string, value interface{})
	// Get the data from store
	Get(key string) interface{}
	// Delete data
	Delete(key string)
	// Destroy the session and delete all related stored data
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
	if mw.cfg.Filter != nil && mw.cfg.Filter(c) {
		c.Next()
		return
	}

	store := mw.core.Get(c)
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

func New(config ...Config) func(*fiber.Ctx) {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	core := fbSession.New(cfg.StoreConfig)
	mw := &sessionMiddleware{cfg: cfg, core: core}
	return mw.handler
}
