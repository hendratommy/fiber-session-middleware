package main

import (
	sessionMW "fiber-session-middleware"
	"github.com/gofiber/fiber"
	"github.com/gofiber/session"
)

func main() {
	app := fiber.New()

	middleware := sessionMW.New(sessionMW.Config{
		StoreConfig: session.Config{
			Lookup:     "cookie:X-Sess-ID",
		},
	})

	app.Use(middleware)

	app.Get("/get", func(c *fiber.Ctx) {
		session := c.Locals("session").(sessionMW.Session)

		c.JSON(fiber.Map{
			"firstName": session.Get("firstName"),
			"lastName":  session.Get("lastName"),
		})
	})

	app.Get("/set", func(c *fiber.Ctx) {
		session := c.Locals("session").(sessionMW.Session)
		session.Set("firstName", "John")
		session.Set("lastName", "Doe")
		c.SendStatus(fiber.StatusNoContent)
	})

	app.Get("/delete", func(c *fiber.Ctx) {
		session := c.Locals("session").(sessionMW.Session)
		session.Delete("lastName")
		c.SendStatus(fiber.StatusNoContent)
	})

	app.Get("/logout", func(c *fiber.Ctx) {
		session := c.Locals("session").(sessionMW.Session)
		session.Destroy()
		c.SendStatus(fiber.StatusNoContent)
	})

	app.Listen(8080)
}
