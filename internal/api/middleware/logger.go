package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Logger returns a middleware that logs HTTP requests
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Continue chain
		err := c.Next()

		// After request
		stop := time.Now()
		latency := stop.Sub(start)

		// Log using Fiber's logger
		log.Infow("Request",
			"timestamp", stop.Format("2006/01/02 - 15:04:05"),
			"status", c.Response().StatusCode(),
			"latency", latency,
			"ip", c.IP(),
			"method", c.Method(),
			"path", c.Path(),
			"handler", c.Route().Name,
		)

		return err
	}
}
