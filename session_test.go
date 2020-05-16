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

func httpTest(t *testing.T, app *fiber.App, url string, header map[string]string, expectedStatus int, expectedResp map[string]interface{}) http.Header {
	req := httptest.NewRequest("GET", url, nil)
	if header != nil && len(header) > 0 {
		for k := range header {
			req.Header.Set(k, header[k])
		}
	}

	if resp, err := app.Test(req); err != nil {
		t.Error("failed to test fiber.App", err)
	} else {
		assert.Equal(t, expectedStatus, resp.StatusCode)

		if expectedResp != nil || len(expectedResp) != 0 {
			b := make(map[string]interface{})
			if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
				t.Error("failed to decode response", err)
			} else {
				assert.Equal(t, expectedResp, b)
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
			if session.Get("firstName") != nil && session.Get("lastName") != nil {
				c.JSON(fiber.Map{
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

//func TestSession_New(t *testing.T) {
//	app := fiber.New()
//	app.Use(New())
//
//	registerSetHandler(app, "set")
//
//	h := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)
//	assert.NotZero(t, h.Get("Set-Cookie"))
//}
//
//func TestSession_App(t *testing.T) {
//	const sessId = "X-Sess-ID"
//	app := fiber.New()
//	middleware := New(Config{
//		StoreConfig: fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)},
//	})
//	app.Use(middleware)
//
//	registerGetHandler(app, "get")
//	registerSetHandler(app, "set")
//
//	httpTest(t, app, "/get", nil, fiber.StatusNotFound, nil)
//	h := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)
//	httpTest(t, app, "/get", map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
//		"firstName": "John",
//		"lastName": "Doe",
//	})
//
//	h2 := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)
//
//	assert.NotEqual(t, h.Get(sessId), h2.Get(sessId))
//}
//
//func TestSession_App_Path(t *testing.T) {
//	const sessId = "X-Sess-ID"
//	app := fiber.New()
//	middleware := New(Config{
//		StoreConfig: fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)},
//	})
//	app.Use("/enabled", middleware)
//
//	registerGetHandler(app, "get")
//	registerSetHandler(app, "set")
//	registerGetHandler(app, "enabled", "get")
//	registerSetHandler(app, "enabled", "set")
//
//	httpTest(t, app, "/get", nil, fiber.StatusNotFound, nil)
//	h := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)
//	assert.Zero(t, h.Get(sessId))
//
//	h2 := httpTest(t, app, "/enabled/set", nil, fiber.StatusNoContent, nil)
//	httpTest(t, app, "/enabled/get", map[string]string{sessId: h2.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
//		"firstName": "John",
//		"lastName": "Doe",
//	})
//}

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

	httpTest(t, app, "/get", nil, fiber.StatusNotFound, nil)
	h := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)
	assert.Zero(t, h.Get(sessId))

	h2 := httpTest(t, app, "/enabled/set", nil, fiber.StatusNoContent, nil)
	httpTest(t, app, "/enabled/get", map[string]string{sessId: h2.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
		"firstName": "John",
		"lastName": "Doe",
	})
}