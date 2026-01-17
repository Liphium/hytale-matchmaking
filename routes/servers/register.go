package servers_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type RegisterServerRequest struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Game       string `json:"game"`
	MaxPlayers int    `json:"max_players"`
}

type RegisterServerResponse struct {
	ID           int    `json:"id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UUID         string `json:"uuid"`
}

// Endpoint: /api/servers/register
func registerServer(c *fiber.Ctx) error {
	var req RegisterServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	info, ok := service.CreateNewServer(req.Game)
	if !ok {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	info.Mutex.Lock()
	defer info.Mutex.Unlock()
	return c.JSON(RegisterServerResponse{
		ID:           info.Id,
		AccessToken:  info.Token.AccessToken,
		RefreshToken: info.Token.RefreshToken,
		UUID:         info.Token.UUID,
	})
}
