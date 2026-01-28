package service

import (
	"slices"
	"sync"
)

// States for matches
const (
	MatchStateAvailable = "available"
	MatchStateAccepting = "accepting"
	MatchStateFull      = "full"
	MatchStateEnd       = "end" // State to mark the match for deletion
)

type Match struct {
	Mutex      *sync.RWMutex
	ID         int      // Unique id (by server)
	Server     int      // What server the match is on
	State      string   // The current state of the match
	Game       string   // The gamemode the match is in
	Players    []string // List of player ids in the match
	TokenStore []string // List of tokens that can still be used
}

// Locks the mutex
func (m *Match) CanBeJoined() bool {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	return m.canBeJoinedNoMutex()
}

func (m *Match) canBeJoinedNoMutex() bool {
	return m.State == MatchStateAccepting && len(m.TokenStore) > 0
}

// Tries to add a player to the match (returns false if it didn't work)
func (m *Match) AddPlayerIfPossible(id string) (string, bool) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if m.canBeJoinedNoMutex() {
		m.Players = append(m.Players, id)

		// Remove the first token and remove it
		token := m.TokenStore[0]
		m.TokenStore = slices.Delete(m.TokenStore, 0, 1)

		return token, true
	}
	return "", false
}

// Remove all players in the match from the system
func (m *Match) deleteAllPlayers() {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	for _, player := range m.Players {
		go DeletePlayer(player, nil) // In a goroutine to make sure no mutex shit happens
	}
}

type MatchRegistry struct {
	Game             string
	Mutex            *sync.RWMutex
	currentlyFilling *Match
	available        []*Match
}

func (mr *MatchRegistry) GetMatch(id int) (*Match, bool) {
	mr.Mutex.RLock()
	defer mr.Mutex.RUnlock()

	i := slices.IndexFunc(mr.available, func(m *Match) bool {
		return m.ID == id
	})
	if i < 0 {
		return nil, false
	}
	return mr.available[i], true
}

// Add a match to the registry
func (mr *MatchRegistry) AddMatch(match *Match) {
	mr.Mutex.Lock()
	defer mr.Mutex.Unlock()

	// Set currently filling when there isn't one
	if mr.currentlyFilling == nil {

		// We use the write mutex to make sure the state isn't changed by something else, cause when it's changed it will be sorted too
		match.Mutex.Lock()
		if match.State == MatchStateAccepting {
			mr.currentlyFilling = match
		}
		match.Mutex.Unlock()
	}

	// Add the match to the regular list
	mr.available = append(mr.available, match)
}

// nil if there isn't any match that is currently available
func (mr *MatchRegistry) getAvailableMatch() *Match {
	mr.Mutex.RLock()
	if mr.currentlyFilling != nil && mr.currentlyFilling.CanBeJoined() {
		return mr.currentlyFilling
	}
	mr.Mutex.RUnlock()

	// If there wasn't a currently filling match, we need to determine a new one
	mr.setBestFillingMatch()

	mr.Mutex.RLock()
	defer mr.Mutex.RUnlock()
	return mr.currentlyFilling
}

// Helper function for setting the best match to fill up next
func (mr *MatchRegistry) setBestFillingMatch() {

	// Clean to make sure no shit happens
	mr.cleanup()

	mr.Mutex.Lock()
	defer mr.Mutex.Unlock()
	mr.currentlyFilling = nil

	currentSize := -1
	for _, match := range mr.available {
		if !match.CanBeJoined() {
			continue
		}

		// Set as currently filling if more players than current or nil
		match.Mutex.RLock()
		if mr.currentlyFilling == nil || currentSize < len(match.Players) {
			mr.currentlyFilling = match
			currentSize = len(match.Players)
		}
		match.Mutex.RUnlock()
	}
}

// Helper function for cleaning up the match registry
func (mr *MatchRegistry) cleanup() {
	mr.Mutex.Lock()
	mr.available = slices.DeleteFunc(mr.available, func(m *Match) bool {
		rem := mr.shouldBeRemoved(m)
		if rem && mr.currentlyFilling == m {
			mr.currentlyFilling = nil
		}
		if rem {
			m.deleteAllPlayers()
		}
		return rem
	})
	mr.Mutex.Unlock()
}

func (mr *MatchRegistry) shouldBeRemoved(match *Match) bool {
	match.Mutex.RLock()
	defer match.Mutex.RUnlock()

	return match.State == MatchStateEnd
}
