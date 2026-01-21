package servers_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {

	// Require the credential as a Header
	router.Use(service.DefaultAuthMiddleware())

	router.Post("/set_access_token", setToken)
	router.Post("/renew", renewServer)
	router.Post("/register", registerServer)
}
