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
	Mutex       *sync.RWMutex
	ID          string // Unique id (by server)
	Server      int    // What server the match is on
	State       string // The current state of the match
	Game        string // The gamemode the match is in
	PlayerCount int    // The current player count
	MaxPlayers  int    // The maximum amount of players that can fit into the match
}

// Locks the mutex
func (m *Match) CanBeJoined() bool {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	return m.canBeJoinedNoMutex()
}

func (m *Match) canBeJoinedNoMutex() bool {
	return m.State == MatchStateAccepting && m.PlayerCount+1 <= m.MaxPlayers
}

// Tries to lock in a player count increment for the match (returns false if it didn't work)
func (m *Match) IncrementPlayerCount() bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if m.canBeJoinedNoMutex() {
		m.PlayerCount += 1
		return true
	}
	return false
}

type MatchRegistry struct {
	Game             string
	Mutex            *sync.RWMutex
	currentlyFilling *Match
	available        []*Match
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

	// Another call to this function might have set it while we were blocking
	if mr.currentlyFilling != nil {
		return
	}

	for _, match := range mr.available {
		if !match.CanBeJoined() {
			continue
		}

		mr.Mutex.RLock()
		defer mr.Mutex.RUnlock()

		// Set as currently filling if more players than current or nil
		if mr.currentlyFilling == nil || mr.currentlyFilling.PlayerCount < match.PlayerCount {
			mr.currentlyFilling = match
		}
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
		return rem
	})
	mr.Mutex.Unlock()
}

func (mr *MatchRegistry) shouldBeRemoved(match *Match) bool {
	match.Mutex.RLock()
	defer match.Mutex.RUnlock()

	return match.State == MatchStateEnd
}
