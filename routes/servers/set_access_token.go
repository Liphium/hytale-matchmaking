package servers_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type SetTokenRequest struct {
	Id          int    `json:"id"`
	AccessToken string `json:"access_token"`
}

// Endpoint: /api/servers/set_access_token
func setToken(c *fiber.Ctx) error {
	var req SetTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	service.ReplaceAccessTokenForServer(req.Id, req.AccessToken)
	return c.SendStatus(fiber.StatusOK)
}
