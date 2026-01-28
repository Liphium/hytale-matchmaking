package service

import (
	"log"
	"slices"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

const PlayerTokenTimeout = 20 * time.Second

type CachedPlayer struct {
	Id     string
	Server int
}

// Player -> Server id (to make sure people don't join twice with the same account)
var PlayerCache *ristretto.Cache[string, CachedPlayer]

func init() {
	var err error
	PlayerCache, err = ristretto.NewCache(&ristretto.Config[string, CachedPlayer]{
		MaxCost:     10_000,      // Maximum 10.000 stored items
		NumCounters: 10_000 * 10, // 10x what we want to store
		BufferItems: 64,          // Read description of field

		OnEvict: func(item *ristretto.Item[CachedPlayer]) {

			// Cleanup player
			go DeletePlayer(item.Value.Id, nil)
		},
	})
	if err != nil {
		log.Fatalln("couldn't create cache:", err)
	}
}

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
	_, ok := PlayerCache.Get(account)
	return ok
}

// nil if no match has available slots (returns the token and server id if success)
func CreatePlayerIfPossible(game string, account string) (string, int, bool) {
	mr, ok := GetMatchRegistry(game)
	if !ok {
		return "", 0, false
	}

	// Find an available match (loops until there really isn't any slot available)
	var playerToken string = ""
	var match *Match
	for {
		match = mr.getAvailableMatch()
		if match == nil {
			return "", 0, false
		}

		if token, ok := match.AddPlayerIfPossible(account); ok {
			playerToken = token
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
		Token:     playerToken,
		Confirmed: false,
	}
	if !addPlayer(match.Server, account, player, true) {
		return "", 0, false
	}
	return player.Token, player.Server, true
}

// Make sure a player token is actually valid (returns true and matchId if the token has successfully been confirmed)
func ConfirmPlayerToken(server int, account string, token string) (int, bool) {

	// Make sure the player is actually valid
	player, ok := getPlayer(account)
	if !ok || player.Confirmed || player.Token != token {
		return 0, false
	}

	player.Mutex.RLock()

	// Validate the match
	match, ok := GetMatchFromServer(server, player.Match)
	if !ok {
		player.Mutex.RUnlock()
		return 0, false
	}

	match.Mutex.RLock()
	defer match.Mutex.RUnlock()

	// Make sure the player has actually been accepted for the match
	if !slices.Contains(match.Players, account) {
		player.Mutex.RUnlock()
		return 0, false
	}

	player.Mutex.RUnlock()
	player.Mutex.Lock()
	defer player.Mutex.Unlock()

	player.Confirmed = true
	addPlayer(server, account, player, false) // Add to make sure they don't get removed by the timeout anymore
	return player.Match, true
}

// Helper function for adding a player to the cache
func addPlayer(server int, account string, player *PlayerInfo, timeout bool) bool {
	info, ok := serverCache.Get(server)
	if !ok {
		return false
	}

	info.Players.Store(account, player)
	if timeout {
		PlayerCache.SetWithTTL(account, CachedPlayer{
			Id:     account,
			Server: server,
		}, 1, PlayerTokenTimeout)
	} else {
		PlayerCache.Set(account, CachedPlayer{
			Id:     account,
			Server: server,
		}, 1)
	}
	PlayerCache.Wait()
	return true
}

// Helper function for getting a player by account id
func getPlayer(account string) (*PlayerInfo, bool) {
	player, ok := PlayerCache.Get(account)
	if !ok {
		return nil, false
	}

	return getPlayerFromCached(player)
}

// Helper function for getting a player by account id
func getPlayerFromCached(player CachedPlayer) (*PlayerInfo, bool) {
	info, ok := serverCache.Get(player.Server)
	if !ok {
		return nil, false
	}

	pObj, ok := info.Players.Load(player.Id)
	if !ok {
		return nil, false
	}
	return pObj.(*PlayerInfo), true
}

// Helper function for deleting a player from everywhere they leave a trace (set the player info if deleted straight from the cache)
func DeletePlayer(account string, cached *CachedPlayer) {
	if cached == nil {
		var ok bool
		obj, ok := PlayerCache.Get(account)
		if !ok {
			return // Already deleted
		}
		cached = &obj
	}

	info, ok := getPlayerFromCached(*cached)
	if !ok {
		return
	}

	info.Mutex.RLock()
	defer info.Mutex.RUnlock()

	// Delete the player from the server
	srv, ok := serverCache.Get(info.Server)
	if ok {
		srv.Players.Delete(account)

		// Delete the player from the match they were in
		m, ok := GetMatchFromServer(info.Server, info.Match)
		if ok {
			m.Mutex.Lock()
			defer m.Mutex.Unlock()

			// Remove the actual account from the match's player list
			m.Players = slices.DeleteFunc(m.Players, func(p string) bool {
				return p == account
			})

			// Make their token available again
			m.TokenStore = append(m.TokenStore, info.Token)
		}
	}

	PlayerCache.Del(account)
	PlayerCache.Wait()
}
