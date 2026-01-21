package players_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type ConfirmPlayerRequest struct {
	Server int    `json:"server"`
	Player string `json:"player"`
	Token  string `json:"token"`
}

type ConfirmPlayerResponse struct {
	Match int `json:"match"`
}

// Route: POST /api/players/confirm
func ConfirmPlayer(c *fiber.Ctx) error {
	var req ConfirmPlayerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Confirm the player token and return the match when it worked
	match, ok := service.ConfirmPlayerToken(req.Server, req.Player, req.Token)
	if !ok {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.JSON(ConfirmPlayerResponse{
		Match: match,
	})
}
