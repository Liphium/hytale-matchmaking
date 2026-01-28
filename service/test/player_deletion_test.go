package service_test

import (
	"testing"

	"github.com/Liphium/hytale-matchmaking/service"
	"github.com/stretchr/testify/assert"
)

func TestPlayerDeletion(t *testing.T) {
	service.ResetAll()

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
	assert.True(t, service.SetMatchState(serverId, matchId, service.MatchStateAccepting))

	t.Run("token gets added back when deleted", func(t *testing.T) {
		token, server, ok := service.CreatePlayerIfPossible(game, "test")
		assert.True(t, ok)
		assert.Equal(t, "test", token)
		assert.Equal(t, 1, server)

		// Delete the player and make sure their token gets added back to the match
		service.DeletePlayer("test", nil)

		match, ok := service.GetMatchFromServer(server, matchId)
		assert.True(t, ok)
		assert.Equal(t, 1, len(match.TokenStore))
		assert.Equal(t, "test", match.TokenStore[0])
	})

	// This is required for the OnEvict callback from the cache itself
	t.Run("works after deletion from the cache", func(t *testing.T) {
		token, server, ok := service.CreatePlayerIfPossible(game, "test")
		assert.True(t, ok)
		assert.Equal(t, "test", token)
		assert.Equal(t, 1, server)

		// Delete from the cache and then try normal deletion (like the callback would)
		service.PlayerCache.Del("test")
		service.DeletePlayer("test", &service.CachedPlayer{
			Id:     "test",
			Server: serverId,
		})

		match, ok := service.GetMatchFromServer(server, matchId)
		assert.True(t, ok)
		assert.Equal(t, 1, len(match.TokenStore))
		assert.Equal(t, "test", match.TokenStore[0])
	})
}
