package control_routes

import (
	"github.com/Liphium/hytale-matchmaking/util"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {

	// Require the credential as a query parameter
	router.Use(func(c *fiber.Ctx) error {

		cred := c.Query("credential", "-")
		if cred != util.GetCredential() {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	})

	router.Get("/add_new", addNewToken)
}
