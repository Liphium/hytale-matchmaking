package servers_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type SetStateRequest struct {
	ID    int    `json:"id"`
	State string `json:"state"`
}

// Endpoint: /api/servers/set_state
func setState(c *fiber.Ctx) error {
	var req SetStateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	service.SetServerState(req.ID, req.State)
	return c.SendStatus(fiber.StatusOK)
}
