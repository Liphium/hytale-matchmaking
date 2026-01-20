package service

import (
	"slices"
	"sync"

	"github.com/Liphium/hytale-matchmaking/util"
)

// Player -> Server id (to make sure people don't join twice with the same account)
var playerCache = &sync.Map{}

type PlayerInfo struct {
	Mutex   *sync.RWMutex
	Account string
	Server  int
	Match   int

	// For actual join behavior
	Token     string
	Confirmed bool
}

// Check if an account is in a match or the queue for one
func IsOnServerOrWaiting(account string) bool {
	_, ok := playerCache.Load(account)
	return ok
}

// nil if no match has available slots
func CreatePlayerIfPossible(game string, account string) *PlayerInfo {
	mr, ok := getMatchRegistry(game)
	if !ok {
		return nil
	}

	// Find an available match (loops until there really isn't any slot available)
	match := mr.getAvailableMatch()
	for {
		match = mr.getAvailableMatch()
		if match == nil {
			return nil
		}

		if match.AddPlayerIfPossible(account) {
			break
		}
	}

	match.Mutex.RLock()
	defer match.Mutex.RUnlock()

	player := &PlayerInfo{
		Mutex:     &sync.RWMutex{},
		Account:   account,
		Server:    match.Server,
		Match:     match.ID,
		Token:     util.GenerateToken(64),
		Confirmed: false,
	}
	if !addPlayer(match.Server, account, player) {
		return nil
	}
	return player
}

// Make sure a player token is actually valid (returns true if the token has successfully been confirmed)
func ConfirmPlayerToken(server int, match int, account string, token string) bool {
	return false
}

// Helper function for adding a player to the cache
func addPlayer(server int, account string, player *PlayerInfo) bool {
	info, ok := serverCache.Get(server)
	if !ok {
		return false
	}

	info.Players.Store(account, player)
	playerCache.Store(account, server)
	return true
}

// Helper function for getting a player by account id
func getPlayer(account string) (*PlayerInfo, bool) {
	obj, ok := playerCache.Load(account)
	if !ok {
		return nil, false
	}
	server := obj.(int)

	info, ok := serverCache.Get(server)
	if !ok {
		return nil, false
	}

	pObj, ok := info.Players.Load(account)
	if !ok {
		return nil, false
	}
	return pObj.(*PlayerInfo), true
}

// Helper function for deleting a player from everywhere they leave a trace
func deletePlayer(account string) {
	info, ok := getPlayer(account)
	if ok {
		info.Mutex.RLock()
		defer info.Mutex.RUnlock()

		// Delete the player from the server
		srv, ok := serverCache.Get(info.Server)
		if ok {
			srv.Players.Delete(account)

			// Delete the player from the match they were in
			m, ok := getMatchFromServer(info.Server, info.Match)
			if ok {
				m.Mutex.Lock()
				defer m.Mutex.Unlock()

				m.Players = slices.DeleteFunc(m.Players, func(p string) bool {
					return p == account
				})
			}
		}
	}
	playerCache.Delete(account)
}
