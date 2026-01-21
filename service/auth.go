package service

import (
	"github.com/Liphium/hytale-matchmaking/util"
	"github.com/gofiber/fiber/v2"
)

func DefaultAuthMiddleware() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		cred := c.Get("Credential", "-")
		if cred != util.GetCredential() {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	}
}
