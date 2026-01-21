package service

func ResetAll() {
	playerCache.Clear()
	serverCache.Clear()
	gameCache.Clear()
	tokensMap.Clear()
}
