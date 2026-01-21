package players_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type QueuePlayerRequest struct {
	Player string `json:"player"`
	Game   string `json:"game"`
}

type QueuePlayerResponse struct {
	Address string `json:"address"` // Address of the server (e.g. liphium.com or 127.0.0.1)
	Port    int    `json:"port"`
	Token   string `json:"token"`
}

// Route: POST /api/players/queue
func QueuePlayer(c *fiber.Ctx) error {
	var req QueuePlayerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	token, server, ok := service.CreatePlayerIfPossible(req.Game, req.Player)
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}

	address, port, ok := service.GetServerDetails(server)
	if !ok {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(QueuePlayerResponse{
		Address: address,
		Port:    port,
		Token:   token,
	})
}
