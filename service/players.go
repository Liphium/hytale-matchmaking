package service

import "sync"

// Player -> Server id (to make sure people don't join twice with the same account)
var playerCache = &sync.Map{}

type PlayerInfo struct {
	Mutex   *sync.RWMutex
	Account string
	Server  int
	Match   int

	// For actual join behavior
	Token  string
	Joined bool
}

// nil if no match has available slots
func AddPlayerToken(server int, game string, account string) *PlayerInfo {
	return nil
}
