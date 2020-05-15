package fiber_session_middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber"
	fbSession "github.com/gofiber/session"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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

func TestSession(t *testing.T) {
	const sessId = "X-Sess-ID"
	app := fiber.New()
	middleware := New(fbSession.Config{Lookup: fmt.Sprintf("header:%s", sessId)})
	app.Use(middleware)

	app.Get("/get", func(c *fiber.Ctx) {
		session := c.Locals("session").(Session)

		if session.Get("firstName") == nil && session.Get("lastName") == nil {
			c.SendStatus(fiber.StatusNotFound)
		} else {
			c.JSON(fiber.Map{
				"firstName": session.Get("firstName"),
				"lastName":  session.Get("lastName"),
			})
		}
	})
	app.Get("/set", func(c *fiber.Ctx) {
		session := c.Locals("session").(Session)
		session.Set("firstName", "John")
		session.Set("lastName", "Doe")
		c.SendStatus(fiber.StatusNoContent)
	})

	httpTest(t, app, "/get", nil, fiber.StatusNotFound, nil)
	h := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)
	httpTest(t, app, "/get", map[string]string{sessId: h.Get(sessId)}, fiber.StatusOK, map[string]interface{}{
		"firstName": "John",
		"lastName": "Doe",
	})

	h2 := httpTest(t, app, "/set", nil, fiber.StatusNoContent, nil)

	assert.NotEqual(t, h.Get(sessId), h2.Get(sessId))
}
