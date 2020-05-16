# fiber-session-middleware
A session `gofiber/fiber` middleware based on `gofiber/session`

## Install
```
go get -u github.com/gofiber/fiber
go get -u github.com/gofiber/session
go get -u github.com/hendratommy/fiber-session-middleware
```

## Example
```go
package main

import (
	sessionMW "fiber-session-middleware"
	"github.com/gofiber/fiber"
)

func main() {
	app := fiber.New()

	middleware := sessionMW.New()

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
```

The `New` function accept `Config` struct which accept `session.Config` to configure the underlying `gofiber/session`.
The middleware will take care of making the `Session` available for your handler as `ctx.Locals` with `session` as the 
key. The middleware also handle the `Save` operation to persist your data, so you don't have to deal with it manually
for every handlers that use `session`.

For example, to configure it using `RedisProvider`

```go
package main

import (
	sessionMW "github.com/github.com/hendratommy/fiber-session-middleware"
	"github.com/gofiber/fiber"
	"github.com/gofiber/session"
	"github.com/gofiber/session/provider/redis"
	"time"
)

func main() {
    app := fiber.New()
    
    redisProvider := redis.New(redis.Config{
        KeyPrefix:   "session",
        Addr:        "localhost:6379",
        DB:          1,
        PoolSize:    8,
        IdleTimeout: 30 * time.Second,
    })
    middleware := sessionMW.New(sessionMW.Config{
        StoreConfig: session.Config{
            Lookup:     "header:X-Sess-ID",
            Provider:   redisProvider,
            Expiration: 1 * time.Hour,
        },
    })
    app.Use(middleware)
}
```

## API Reference

### Config
| Property | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| Filter | `func(*fiber.Ctx) bool` | Defines a function to skip middleware | `nil`
| StoreConfig | `session.Config` | `Config` struct from [gofiber/session](https://github.com/gofiber/session) | `session.Config{}`