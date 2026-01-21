package servers_routes

import (
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/gofiber/fiber/v2"
)

type RegisterServerRequest struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
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

	// Find a valid token
	token, ok := service.GetFreeToken()
	if !ok {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	token.Mutex.Lock()
	defer token.Mutex.Unlock()

	service.CreateServer(token.Id, req.IP, req.Port)
	return c.JSON(RegisterServerResponse{
		ID:           token.Id,
		AccessToken:  token.Token.AccessToken,
		RefreshToken: token.Token.RefreshToken,
		UUID:         token.Token.UUID,
	})
}
