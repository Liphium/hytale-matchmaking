package service

import "sync"

// Game (string) -> *MatchRegistry
var gameCache = &sync.Map{}

// Returns whether or not the match could be registered (state and stuff will be adjusted)
func AddMatch(server int, match *Match) bool {
	info, ok := serverCache.Get(server)
	if !ok {
		return false
	}

	// Make sure server stuff can be read
	info.Mutex.RLock()
	defer info.Mutex.RUnlock()

	if match.Mutex == nil {
		match.Mutex = &sync.RWMutex{}
	}

	// Save the match to the cache
	match.Mutex.Lock()
	gameCp := match.Game
	match.Server = info.TokenId
	match.State = MatchStateAvailable
	info.Matches.Store(match.ID, match)
	match.Mutex.Unlock()

	// Add to the game
	addMatchToGame(gameCp, match)

	return true
}

func addMatchToGame(game string, match *Match) {
	obj, ok := gameCache.Load(game)
	if !ok {
		obj = &MatchRegistry{
			Game:             game,
			Mutex:            &sync.RWMutex{},
			currentlyFilling: match,
			available:        []*Match{match},
		}
	}

	registry := obj.(*MatchRegistry)
	registry.AddMatch(match)
}

// Returns false when it didn't work
func SetMatchState(server int, matchId int, state string) bool {
	match, ok := getMatchFromServer(server, matchId)
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

// Helper function for quickly getting a match
func getMatchFromServer(server int, match int) (*Match, bool) {
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
