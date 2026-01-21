package players_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {

	// Require the credential as a Header
	router.Use(service.DefaultAuthMiddleware())

	router.Post("/confirm", ConfirmPlayer)
	router.Post("/queue", QueuePlayer)
}
