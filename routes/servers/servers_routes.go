package servers_routes

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

	router.Post("/set_access_token", setToken)
	router.Post("/renew", renewServer)
	router.Post("/register", registerServer)
	router.Post("/set_state", setState)
}
