# Hytale is very cool

Because the Hytale server is really well made, they don't have an async login event...

This means the following things have to be done:

- Match servers maintain a token store for them to be able to check tokens without API requests
	- Along with the max players they should send tokens for the players that they generated
- Write a redirect server in Go to be able to properly redirect logins???
