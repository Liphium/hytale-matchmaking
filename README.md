# Match-making for Hytale servers

> [!WARNING]
> This project is in early stages, expect lots of bugs and changes in the coming weeks. We're still actively figuring out the proper architecture for a minigame network in Hytale ourselves.

When making a Hytale server currently, one of the hardest things is to structure it since there is basically no reference. This is our attempt of a nice Hytale minigame server architecture.

This repository is the core of the match-making system we're currently designing, it's basically a API that can handle player management and matches all under one roof, in one central place, read more below.

## Features

> [!NOTE]
> The plugins actually making this system fully functional are still not public. We will publish them in the coming weeks.

- Let servers automatically authenticate themselves using a central token storage
- Match-making across multiple Game modes with the Game server in full control
  - API for your plugin to control matchmaking
  - Automatically get the server with the lowest player count to send players to
- Redirect servers to automatically connect players to your network with safety in mind
- No proxy required (The entire system uses Hytale redirects)
- Automatic detection of servers going offline (will not send notifications, but not redirect players there)

### Planned

- Automatic E-Mail notifications when not enough tokens are available to the service
- Monitoring for players, server health and match health
- Spectator support: Enable players to join as spectators
- Integration with Agones and Kubernetes
