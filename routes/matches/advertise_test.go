package matches_routes_test

import (
	"testing"

	matches_routes "github.com/Liphium/hytale-matchmaking/routes/matches"
	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/Liphium/hytale-matchmaking/util"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func TestAdvertise(t *testing.T) {
	service.ResetAll()

	const (
		id     = 1
		server = "localhost"
		port   = 3000
		game   = "battle"
	)

	assert.True(t, service.CreateServer(id, server, port))

	matchToCreate := service.MatchCreate{
		ID:         1,
		Game:       game,
		MaxPlayers: 1,
	}

	t.Run("match can be advertised", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		_, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(matches_routes.AdvertiseMatchRequest{
				Server: id,
				Match:  matchToCreate,
			}).
			Post(util.DefaultPath("/api/matches/advertise"))
		assert.Nil(t, err)

		// Make sure the correct match has been created in the service
		reg, ok := service.GetMatchRegistry(game)
		assert.True(t, ok)
		match, ok := reg.GetMatch(id)
		assert.True(t, ok)
		assert.Equal(t, matchToCreate.ID, match.ID)
		assert.Equal(t, matchToCreate.Game, match.Game)
		assert.Equal(t, matchToCreate.MaxPlayers, match.MaxPlayers)
		assert.Equal(t, 0, len(match.Players))
		assert.Equal(t, service.MatchStateAvailable, match.State)
		assert.Equal(t, id, match.Server)
	})

	t.Run("can't advertise same match", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		res, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(matches_routes.AdvertiseMatchRequest{
				Server: id,
				Match:  matchToCreate,
			}).
			Post(util.DefaultPath("/api/matches/advertise"))
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, res.StatusCode())
	})

	t.Run("can't advertise match on non-existent server", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		res, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(matches_routes.AdvertiseMatchRequest{
				Server: 5,
				Match:  matchToCreate,
			}).
			Post(util.DefaultPath("/api/matches/advertise"))
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, res.StatusCode())
	})
}
