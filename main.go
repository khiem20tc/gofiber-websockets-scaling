package main

import (
	"encoding/json"
	"gofiber-ws/config"
	"gofiber-ws/redis"
	"gofiber-ws/websocket"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func main() {

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(cors.New())

	// Initialize default config (Assign the middleware to /metrics)
	app.Get("/metrics", monitor.New())
	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// Initialize default config
	app.Use(logger.New(logger.Config{}))

	// Initialize redis
	redis.Init()

	// Handle websocket
	app.Get("/ws/:session", websocket.UpgradeWebsocket, websocket.HandleWebSocket)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	if err := app.Listen(":" + config.Env("PORT", "3002")); err != nil {
		log.Fatal(err)
	}
}
