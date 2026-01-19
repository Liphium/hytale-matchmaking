package service

import "sync"

// Player -> Server id (to make sure people don't join twice with the same account)
var playerCache = &sync.Map{}

type PlayerInfo struct {
	Account string
	Server  int
	Token   string
	Joined  bool
}

// nil if no server has available slots
func AddPlayerToken(server int, account string) *PlayerInfo {
	return nil
}
