package service

func ResetAll() {
	PlayerCache.Clear()
	serverCache.Clear()
	gameCache.Clear()
	tokensMap.Clear()
}
