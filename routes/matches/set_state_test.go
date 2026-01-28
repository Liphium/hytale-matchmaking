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

func TestSetState(t *testing.T) {
	service.ResetAll()

	const (
		id     = 1
		server = "localhost"
		port   = 3000
		game   = "battle"
	)

	assert.True(t, service.CreateServer(id, server, port))

	// Create a test match
	created := service.MatchCreate{
		ID:   1,
		Game: game,
	}
	assert.True(t, service.AddMatch(1, created, []string{"test"}))

	t.Run("match state can be changed", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		_, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(matches_routes.MatchSetStateRequest{
				Server: id,
				Match:  created.ID,
				State:  service.MatchStateFull,
			}).
			Post(util.DefaultPath("/api/matches/set_state"))
		assert.Nil(t, err)

		match, ok := service.GetMatchFromServer(id, created.ID)
		assert.True(t, ok)
		assert.Equal(t, service.MatchStateFull, match.State)
	})

	t.Run("end deletes match", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		_, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(matches_routes.MatchSetStateRequest{
				Server: id,
				Match:  created.ID,
				State:  service.MatchStateEnd,
			}).
			Post(util.DefaultPath("/api/matches/set_state"))
		assert.Nil(t, err)

		_, ok := service.GetMatchFromServer(id, created.ID)
		assert.False(t, ok)
	})

	t.Run("state of invalid match can't be changed", func(t *testing.T) {
		client := resty.New()
		defer client.Close()

		res, err := client.R().
			SetHeaders(util.CredentialHeaders()).
			SetBody(matches_routes.MatchSetStateRequest{
				Server: id,
				Match:  67,
				State:  service.MatchStateEnd,
			}).
			Post(util.DefaultPath("/api/matches/set_state"))
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, res.StatusCode())

		_, ok := service.GetMatchFromServer(id, 67)
		assert.False(t, ok)
	})
}
