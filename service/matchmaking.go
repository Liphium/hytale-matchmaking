package service

import (
	"sync"
)

// Game (string) -> *MatchRegistry
var gameCache = &sync.Map{}

type MatchCreate struct {
	ID         int    `json:"id"`          // Unique id (by server)
	Game       string `json:"game"`        // The gamemode the match is in
	MaxPlayers int    `json:"max_players"` // The maximum amount of players that can fit into the match
}

// Returns whether or not the match could be registered (state and stuff will be adjusted)
func AddMatch(server int, data MatchCreate) bool {
	info, ok := serverCache.Get(server)
	if !ok {
		return false
	}

	// Make sure server stuff can be read
	info.Mutex.RLock()
	defer info.Mutex.RUnlock()

	// Make sure the match doesn't already exist
	if _, ok := info.Matches.Load(data.ID); ok {
		return false
	}

	// Initialize the match with the data from the request
	match := &Match{
		Mutex:      &sync.RWMutex{},
		ID:         data.ID,
		Game:       data.Game,
		MaxPlayers: data.MaxPlayers,
		Server:     server,
		State:      MatchStateAvailable,
	}

	// Add to the game
	info.Matches.Store(data.ID, match)
	addMatchToGame(data.Game, match)

	return true
}

func addMatchToGame(game string, match *Match) {
	obj, ok := gameCache.Load(game)
	if !ok {
		obj = &MatchRegistry{
			Game:      game,
			Mutex:     &sync.RWMutex{},
			available: []*Match{},
		}
		gameCache.Store(game, obj)
	}

	registry := obj.(*MatchRegistry)
	registry.AddMatch(match)
}

// Returns false when it didn't work
func SetMatchState(server int, matchId int, state string) bool {
	match, ok := GetMatchFromServer(server, matchId)
	if !ok {
		return false
	}

	match.Mutex.Lock()
	match.State = state
	match.Mutex.Unlock()

	// Delete the match when it ends
	if state == MatchStateEnd {
		server, ok := serverCache.Get(server)
		if ok {
			server.Matches.Delete(matchId)
		}
		return true
	}
	return true
}

// Get the match registry for a game
func GetMatchRegistry(game string) (*MatchRegistry, bool) {
	obj, ok := gameCache.Load(game)
	if !ok {
		return nil, false
	}
	return obj.(*MatchRegistry), true
}

// Helper function for quickly getting a match
func GetMatchFromServer(server int, match int) (*Match, bool) {
	info, ok := serverCache.Get(server)
	if !ok {
		return nil, false
	}

	// Make sure server stuff can be read
	info.Mutex.RLock()
	defer info.Mutex.RUnlock()

	obj, ok := info.Matches.Load(match)
	if !ok {
		return nil, false
	}
	return obj.(*Match), true
}
