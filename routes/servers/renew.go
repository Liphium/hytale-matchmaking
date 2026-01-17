package servers_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type RenewServerRequest struct {
	ID int `json:"id"`
}

// Endpoint: /api/servers/renew
func renewServer(c *fiber.Ctx) error {
	var req RenewServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	service.RefreshServer(req.ID)
	return c.SendStatus(fiber.StatusOK)
}
