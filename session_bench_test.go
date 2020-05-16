package fiber_session_middleware

import (
	"fmt"
	"github.com/gofiber/fiber"
	fbSession "github.com/gofiber/session"
	"github.com/google/uuid"
	"testing"
)

// go test -v ./... -run=^$ -bench=Benchmark -benchmem -count=3

var keys []string

// generate random keys
func init() {
	for i := 0; i < 128; i++ {
		keys = append(keys, uuid.New().String())
	}
}

func registerBenchSetHandler(app *fiber.App) {
	app.Get("/get", func(c *fiber.Ctx) {
		m := fiber.Map{}
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)

			if session.Get(keys[0]) != nil {
				for _, key := range keys {
					m[key] = session.Get(key)
				}
			}
		}
		c.JSON(m)
	})

	app.Get("/set", func(c *fiber.Ctx) {
		m := fiber.Map{}
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)

			for _, key := range keys {
				uid, _ := uuid.NewRandom()
				v := uid.String()
				session.Set(key, v)
				m[key] = v
			}
		}
		c.JSON(m)
	})

	app.Get("/delete", func(c *fiber.Ctx) {
		m := fiber.Map{}
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)

			for i, key := range keys {
				if i >= 40 && i < 60 {
					session.Delete(key)
				} else {
					m[key] = session.Get(key)
				}
			}
		}
		c.JSON(m)
	})

	app.Get("/destroy", func(c *fiber.Ctx) {
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)
			session.Destroy()
		}
		c.SendStatus(fiber.StatusNoContent)
	})
}

func BenchmarkApp(b *testing.B) {
	const sessId = "X-Sess-ID"
	app := fiber.New()
	middleware := New(Config{
		StoreConfig: fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)},
	})
	app.Use(middleware)

	registerBenchSetHandler(app)

	for n := 0; n < b.N; n++ {
		h := httpTest(b, app, nil, fiber.StatusOK, len(keys), "set")
		httpTest(b, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, len(keys), "get")
		httpTest(b, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, len(keys)-20, "delete")
		httpTest(b, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusNoContent, nil, "destroy")
		httpTest(b, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, 0, "get")
	}
}
