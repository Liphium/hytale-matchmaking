package starter

import (
	"os"

	"github.com/Liphium/hytale-matchmaking/routes"
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func Start() {
	godotenv.Load()
	service.LoadTokens()

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	// Setup all the routes
	app.Route("/api", routes.SetupRoutes)

	// Hello world handler for health checks
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Hytale Matchmaking!")
	})

	app.Listen(os.Getenv("LISTEN"))
}
