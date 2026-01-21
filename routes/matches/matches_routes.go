package matches_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {

	// Require the credential as a header
	router.Use(service.DefaultAuthMiddleware())

	router.Post("/advertise", AdvertiseMatch)
	router.Post("/set_state", SetMatchState)
}
