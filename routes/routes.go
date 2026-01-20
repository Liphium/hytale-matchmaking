package routes

import (
	control_routes "github.com/Liphium/hytale-matchmaking/routes/control"
	matches_routes "github.com/Liphium/hytale-matchmaking/routes/matches"
	players_routes "github.com/Liphium/hytale-matchmaking/routes/players"
	servers_routes "github.com/Liphium/hytale-matchmaking/routes/servers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(router fiber.Router) {
	router.Route("/control", control_routes.SetupRoutes)
	router.Route("/servers", servers_routes.SetupRoutes)
	router.Route("/players", players_routes.SetupRoutes)
	router.Route("/matches", matches_routes.SetupRoutes)
}
