package fiber_session_middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber"
	fbSession "github.com/gofiber/session"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func httpTest(t assert.TestingT, app *fiber.App, header map[string]string, expectedStatus int, expectedResp interface{}, paths ...string) http.Header {
	path := fmt.Sprintf("/%s", strings.Join(paths, "/"))

	req := httptest.NewRequest("GET", path, nil)
	if header != nil && len(header) > 0 {
		for k := range header {
			req.Header.Set(k, header[k])
		}
	}

	if resp, err := app.Test(req); err != nil {
		assert.NoError(t, err)
	} else {
		assert.Equal(t, expectedStatus, resp.StatusCode)

		if expectedResp != nil {
			b := make(map[string]interface{})
			if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
				assert.NoError(t, err)
			}

			switch v := expectedResp.(type) {
			case map[string]interface{}:
				assert.Equal(t, v, b)

			case int:
				l := len(b)
				assert.Equal(t, v, l)
			}
		}
		return resp.Header
	}
	return nil
}

func registerGetHandler(app *fiber.App, paths ...string) {
	path := fmt.Sprintf("/%s", strings.Join(paths, "/"))

	app.Get(path, func(c *fiber.Ctx) {
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)
			if session.Get("firstName") != nil || session.Get("lastName") != nil {
				c.JSON(fiber.Map{
					"id":        session.ID(),
					"firstName": session.Get("firstName"),
					"lastName":  session.Get("lastName"),
				})
				return
			}
		}
		c.SendStatus(fiber.StatusNotFound)
	})
}

func registerSetHandler(app *fiber.App, paths ...string) {
	path := fmt.Sprintf("/%s", strings.Join(paths, "/"))

	app.Get(path, func(c *fiber.Ctx) {
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)
			session.Set("firstName", "John")
			session.Set("lastName", "Doe")
		}
		c.SendStatus(fiber.StatusNoContent)
	})
}

func registerDeleteHandler(app *fiber.App, paths ...string) {
	path := fmt.Sprintf("/%s", strings.Join(paths, "/"))

	app.Get(path, func(c *fiber.Ctx) {
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)
			session.Delete("lastName")
		}
		c.SendStatus(fiber.StatusNoContent)
	})
}

func registerDestroyHandler(app *fiber.App, paths ...string) {
	path := fmt.Sprintf("/%s", strings.Join(paths, "/"))

	app.Get(path, func(c *fiber.Ctx) {
		if cs := c.Locals("session"); cs != nil {
			session := cs.(Session)
			session.Destroy()
		}
		c.SendStatus(fiber.StatusNoContent)
	})
}

func TestSession_New(t *testing.T) {
	app := fiber.New()
	app.Use(New())

	registerSetHandler(app, "set")

	h := httpTest(t, app, nil, fiber.StatusNoContent, nil, "set")
	assert.NotZero(t, h.Get("Set-Cookie"))
}

func TestSession_App(t *testing.T) {
	const sessId = "X-Sess-ID"
	app := fiber.New()
	middleware := New(Config{
		StoreConfig: fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)},
	})
	app.Use(middleware)

	registerGetHandler(app, "get")
	registerSetHandler(app, "set")

	httpTest(t, app, nil, fiber.StatusNotFound, nil, "get")
	h := httpTest(t, app, nil, fiber.StatusNoContent, nil, "set")
	httpTest(t, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
		"id":        h.Get(sessId),
		"firstName": "John",
		"lastName":  "Doe",
	}, "get")

	h2 := httpTest(t, app, nil, fiber.StatusNoContent, nil, "set")
	assert.NotEqual(t, h.Get(sessId), h2.Get(sessId))

	registerDeleteHandler(app, "delete")
	httpTest(t, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusNoContent, nil, "delete")
	httpTest(t, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
		"id":        h.Get(sessId),
		"firstName": "John",
		"lastName":  nil,
	}, "get")

	registerDestroyHandler(app, "destroy")
	httpTest(t, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusNoContent, nil, "destroy")
	httpTest(t, app, map[string]string{sessId: h.Get(sessId)}, fiber.StatusNotFound, nil, "get")

}

func TestSession_App_Path(t *testing.T) {
	const sessId = "X-Sess-ID"
	app := fiber.New()
	middleware := New(Config{
		StoreConfig: fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)},
	})
	app.Use("/enabled", middleware)

	registerGetHandler(app, "get")
	registerSetHandler(app, "set")
	registerGetHandler(app, "enabled", "get")
	registerSetHandler(app, "enabled", "set")

	httpTest(t, app, nil, fiber.StatusNotFound, nil, "get")
	h := httpTest(t, app, nil, fiber.StatusNoContent, nil, "set")
	assert.Zero(t, h.Get(sessId))

	h2 := httpTest(t, app, nil, fiber.StatusNoContent, nil, "enabled", "set")
	httpTest(t, app, map[string]string{sessId: h2.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
		"id":        h2.Get(sessId),
		"firstName": "John",
		"lastName":  "Doe",
	}, "enabled", "get")
}

func TestSession_Filtered(t *testing.T) {
	const sessId = "X-Sess-ID"
	app := fiber.New()
	middleware := New(Config{
		Filter: func(c *fiber.Ctx) bool {
			return strings.HasPrefix(c.Path(), "/get") || strings.HasPrefix(c.Path(), "/set")
		},
		StoreConfig: fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)},
	})
	app.Use(middleware)

	registerGetHandler(app, "get")
	registerSetHandler(app, "set")
	registerGetHandler(app, "enabled", "get")
	registerSetHandler(app, "enabled", "set")

	httpTest(t, app, nil, fiber.StatusNotFound, nil, "get")
	h := httpTest(t, app, nil, fiber.StatusNoContent, nil, "set")
	assert.Zero(t, h.Get(sessId))

	h2 := httpTest(t, app, nil, fiber.StatusNoContent, nil, "enabled", "set")
	httpTest(t, app, map[string]string{sessId: h2.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
		"id":        h2.Get(sessId),
		"firstName": "John",
		"lastName":  "Doe",
	}, "enabled", "get")
}
