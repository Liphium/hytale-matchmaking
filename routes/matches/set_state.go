package matches_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type MatchSetStateRequest struct {
	Server int    `json:"server"`
	Match  int    `json:"match"`
	State  string `json:"state"`
}

// Route: POST /api/matches/set_state
func SetMatchState(c *fiber.Ctx) error {
	var req MatchSetStateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !service.SetMatchState(req.Server, req.Match, req.State) {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	return c.SendStatus(fiber.StatusOK)
}
