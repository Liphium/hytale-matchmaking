package servers_routes

// This endpoint serves as a reference for the server to know if it should start a new match. Our match-making service should now know how many servers should be kept alive at a time. If we notice that currently more game servers than needed are hosting matches, we want to shut one down so others can be used.
// In this case, in this endpoint, we should choose the server with the lowest match count and tell it to never start a new match using this endpoint. This system should be able to do this dynamically based on the current conditions and also based on the fact, that we may need to shut down multiple servers.
// When a game server gets told that it can't start new matches and doesn't currently host any, it should shut itself down.
