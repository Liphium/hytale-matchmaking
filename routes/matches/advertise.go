package matches_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type AdvertiseMatchRequest struct {
	Server int                 `json:"server"`
	Match  service.MatchCreate `json:"match"`
}

// Route: POST /api/matches/advertise
func AdvertiseMatch(c *fiber.Ctx) error {
	var req AdvertiseMatchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !service.AddMatch(req.Server, req.Match) {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	return c.SendStatus(fiber.StatusOK)
}
