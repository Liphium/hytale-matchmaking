package players_routes

import (
	"github.com/Liphium/hytale-matchmaking/util"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {

	// Require the credential as a Header
	router.Use(func(c *fiber.Ctx) error {

		cred := c.Get("Credential", "-")
		if cred != util.GetCredential() {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	})

	router.Post("/confirm", ConfirmPlayer)
	router.Post("/queue", QueuePlayer)
}
