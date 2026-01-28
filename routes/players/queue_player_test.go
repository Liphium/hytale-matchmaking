package players_routes_test

import (
	"testing"

	players_routes "github.com/Liphium/hytale-matchmaking/routes/players"
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/Liphium/hytale-matchmaking/util"
	testing_util "github.com/Liphium/hytale-matchmaking/util/testing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func TestPlayerQueuing(t *testing.T) {
	service.ResetAll()

	// Create a test server and test match
	const (
		serverId = 1
		server   = "localhost"
		port     = 3000
		game     = "battle"
		matchId  = 1
	)
	assert.True(t, service.CreateServer(serverId, server, port))
	assert.True(t, service.AddMatch(serverId, service.MatchCreate{
		ID:   matchId,
		Game: game,
	}, []string{"test"}))
	service.SetMatchState(serverId, matchId, service.MatchStateAccepting)

	t.Run("player can be queued", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		res, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(players_routes.QueuePlayerRequest{
				Player: "test",
				Game:   game,
			}).
			Post(util.DefaultPath("/api/players/queue"))
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, res.StatusCode())

		var r players_routes.QueuePlayerResponse
		testing_util.Unmarshal(t, res.Bytes(), &r)

		assert.Equal(t, "test", r.Token)
		assert.Equal(t, server, r.Address)
		assert.Equal(t, port, r.Port)
	})

	t.Run("queueing another player fails", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		res, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(players_routes.QueuePlayerRequest{
				Player: "test2",
				Game:   game,
			}).
			Post(util.DefaultPath("/api/players/queue"))
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotFound, res.StatusCode())
	})

	t.Run("queueing the same player fails", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		res, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(players_routes.QueuePlayerRequest{
				Player: "test",
				Game:   game,
			}).
			Post(util.DefaultPath("/api/players/queue"))
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotFound, res.StatusCode())
	})
}
